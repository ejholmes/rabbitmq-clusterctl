package clusterctl

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
)

// NullMasterController is a MasterController that does nothing.
var NullMasterController = &staticMasterController{
	node: "rabbit@localhost",
}

// MasterController is an interface for setting the rabbitmq master node.
type MasterController interface {
	Master() (node string, err error)
	SetMaster(node string) error
}

type elbClient interface {
	DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DeregisterInstancesFromLoadBalancer(*elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error)
	RegisterInstancesWithLoadBalancer(*elb.RegisterInstancesWithLoadBalancerInput) (*elb.RegisterInstancesWithLoadBalancerOutput, error)
}

type ec2Client interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

// ELBMasterController implements the MasterController interface using an ELB as
// the source of truth.
type ELBMasterController struct {
	// The ID of the ELB to use.
	LoadBalancerName string

	elb elbClient
	ec2 ec2Client
}

func NewELBMasterController(loadBalancerName string) *ELBMasterController {
	s := session.New()
	return &ELBMasterController{
		LoadBalancerName: loadBalancerName,
		elb:              elb.New(s),
		ec2:              ec2.New(s),
	}
}

// Master returns the node name of the current master.
func (c *ELBMasterController) Master() (string, error) {
	hostname, err := c.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("rabbit@%s", hostname), nil
}

var errNoPrivateDNS = errors.New("ec2 instance does not have a PrivateDnsName")

// Hostname returns the private dns name for the ec2 instance.
func (c *ELBMasterController) Hostname() (string, error) {
	id, err := c.InstanceID()
	if err != nil {
		return "", err
	}

	resp, err := c.ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	if err != nil {
		return "", err
	}

	instance := resp.Reservations[0].Instances[0]

	if instance.PrivateDnsName == nil {
		return "", errNoPrivateDNS
	}

	return *instance.PrivateDnsName, nil
}

var (
	errNoLoadBalancer   = errors.New("load balancer does not exist")
	errNoInstances      = errors.New("no instances attached to load balancer")
	errTooManyInstances = errors.New("expected only 1 instance attached to load balancer")
)

// Returns the id of the ec2 instance that is the current master.
func (c *ELBMasterController) InstanceID() (string, error) {
	instances, err := c.instances()
	if err != nil {
		return "", err
	}

	// There should always be 1 instance attached to the load balancer.
	if len(instances) == 0 {
		return "", errNoInstances
	}

	// There should be AT MOST 1 instance attached to the load balancer.
	if len(instances) > 1 {
		return "", errTooManyInstances
	}

	instance := instances[0]
	return *instance.InstanceId, nil
}

func (c *ELBMasterController) instances() ([]*elb.Instance, error) {
	resp, err := c.elb.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(c.LoadBalancerName)},
	})
	if err != nil {
		return nil, err
	}

	// If there are no load balancer descriptions, there must be no load
	// balancer!
	if len(resp.LoadBalancerDescriptions) == 0 {
		return nil, errNoLoadBalancer
	}

	return resp.LoadBalancerDescriptions[0].Instances, nil
}

// http://rubular.com/r/beIh7d7UE3
var nodeRegexp = regexp.MustCompile(`(\S*)@(\S*)`)

// returns the hostname portion of a rabbitmq node.
func nodeHostname(node string) string {
	matches := nodeRegexp.FindStringSubmatch(node)
	return matches[2]
}

// SetMaster sets the node to be the new master.
func (c *ELBMasterController) SetMaster(node string) error {
	hostname := nodeHostname(node)

	id, err := c.instanceWithHostname(hostname)
	if err != nil {
		return err
	}

	if err := c.SetInstance(id); err != nil {
		return err
	}

	return nil
}

// RemoveInstances removes all ec2 instances from the load balancer.
func (c *ELBMasterController) RemoveInstances() error {
	instances, err := c.instances()
	if err != nil {
		return err
	}

	_, err = c.elb.DeregisterInstancesFromLoadBalancer(&elb.DeregisterInstancesFromLoadBalancerInput{
		LoadBalancerName: aws.String(c.LoadBalancerName),
		Instances:        instances,
	})

	return err
}

// SetInstance removes all instances from the load balancer and sets the master
// to the given instance id.
func (c *ELBMasterController) SetInstance(instanceID string) error {
	if err := c.RemoveInstances(); err != nil {
		return err
	}

	_, err := c.elb.RegisterInstancesWithLoadBalancer(&elb.RegisterInstancesWithLoadBalancerInput{
		LoadBalancerName: aws.String(c.LoadBalancerName),
		Instances: []*elb.Instance{
			{InstanceId: aws.String(instanceID)},
		},
	})

	return err
}

const filterPrivateDnsName = "private-dns-name"

func (c *ELBMasterController) instanceWithHostname(hostname string) (string, error) {
	resp, err := c.ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(filterPrivateDnsName),
				Values: []*string{aws.String(hostname)},
			},
		},
	})
	if err != nil {
		return "", err
	}

	instance := resp.Reservations[0].Instances[0]

	return *instance.InstanceId, nil
}

// staticMasterController is a MasterController implementation that manages a
// static master node that never changes.
type staticMasterController struct {
	node string
}

func (c *staticMasterController) Master() (string, error) {
	return c.node, nil
}

// SetMaster sets the node to be the new master.
func (c *staticMasterController) SetMaster(node string) error {
	return errors.New("master node is static and cannot be changed")
}

// syncQueuesMasterController is a MasterController middleware that ensures that
// queues are synchronized on the current node before setting the new master.
type syncQueuesMasterController struct {
	MasterController
	rabbitmqctl rabbitmqctlFunc
}

// SyncQueues wraps the MasterController with middleware to ensure that all
// queues are synchronized before switching to the new master.
func SyncQueues(m MasterController) MasterController {
	return &syncQueuesMasterController{
		MasterController: m,
		rabbitmqctl:      rabbitmqctl,
	}
}

func (c *syncQueuesMasterController) SetMaster(node string) error {
	return errors.New("not implemented")
}
