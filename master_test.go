package clusterctl

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestELBMasterController_Master(t *testing.T) {
	elbClient := new(mockELBClient)
	ec2Client := new(mockEC2Client)
	c := &ELBMasterController{
		LoadBalancerName: "rabbitmq",
		elb:              elbClient,
		ec2:              ec2Client,
	}

	elbClient.On("DescribeLoadBalancers", &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String("rabbitmq")},
	}).Return(&elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
			{
				Instances: []*elb.Instance{
					{InstanceId: aws.String("i-1234")},
				},
			},
		},
	}, nil)

	ec2Client.On("DescribeInstances", &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String("i-1234")},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{
						PrivateDnsName: aws.String("ip-1-2-3-4.ec2.internal"),
					},
				},
			},
		},
	}, nil)

	node, err := c.Master()
	assert.NoError(t, err)
	assert.Equal(t, "rabbit@ip-1-2-3-4.ec2.internal", node)

	elbClient.AssertExpectations(t)
	ec2Client.AssertExpectations(t)
}

func TestELBMasterController_Master_NoPrivateDNS(t *testing.T) {
	elbClient := new(mockELBClient)
	ec2Client := new(mockEC2Client)
	c := &ELBMasterController{
		LoadBalancerName: "rabbitmq",
		elb:              elbClient,
		ec2:              ec2Client,
	}

	elbClient.On("DescribeLoadBalancers", &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String("rabbitmq")},
	}).Return(&elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
			{
				Instances: []*elb.Instance{
					{InstanceId: aws.String("i-1234")},
				},
			},
		},
	}, nil)

	ec2Client.On("DescribeInstances", &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String("i-1234")},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{},
				},
			},
		},
	}, nil)

	_, err := c.Master()
	assert.Equal(t, errNoPrivateDNS, err)

	elbClient.AssertExpectations(t)
	ec2Client.AssertExpectations(t)
}

func TestELBMasterController_Master_NoLoadBalancer(t *testing.T) {
	elbClient := new(mockELBClient)
	ec2Client := new(mockEC2Client)
	c := &ELBMasterController{
		LoadBalancerName: "rabbitmq",
		elb:              elbClient,
		ec2:              ec2Client,
	}

	elbClient.On("DescribeLoadBalancers", &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String("rabbitmq")},
	}).Return(&elb.DescribeLoadBalancersOutput{}, nil)

	_, err := c.Master()
	assert.Equal(t, errNoLoadBalancer, err)

	elbClient.AssertExpectations(t)
	ec2Client.AssertExpectations(t)
}

func TestELBMasterController_Master_NoInstances(t *testing.T) {
	elbClient := new(mockELBClient)
	ec2Client := new(mockEC2Client)
	c := &ELBMasterController{
		LoadBalancerName: "rabbitmq",
		elb:              elbClient,
		ec2:              ec2Client,
	}

	elbClient.On("DescribeLoadBalancers", &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String("rabbitmq")},
	}).Return(&elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
			{},
		},
	}, nil)

	_, err := c.Master()
	assert.Equal(t, errNoInstances, err)

	elbClient.AssertExpectations(t)
	ec2Client.AssertExpectations(t)
}

func TestELBMasterController_Master_TooManyInstances(t *testing.T) {
	elbClient := new(mockELBClient)
	ec2Client := new(mockEC2Client)
	c := &ELBMasterController{
		LoadBalancerName: "rabbitmq",
		elb:              elbClient,
		ec2:              ec2Client,
	}

	elbClient.On("DescribeLoadBalancers", &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String("rabbitmq")},
	}).Return(&elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
			{
				Instances: []*elb.Instance{
					{InstanceId: aws.String("i-1234")},
					{InstanceId: aws.String("i-2345")},
				},
			},
		},
	}, nil)

	_, err := c.Master()
	assert.Equal(t, errTooManyInstances, err)

	elbClient.AssertExpectations(t)
	ec2Client.AssertExpectations(t)
}

func TestELBMasterController_SetMaster(t *testing.T) {
	elbClient := new(mockELBClient)
	ec2Client := new(mockEC2Client)
	c := &ELBMasterController{
		LoadBalancerName: "rabbitmq",
		elb:              elbClient,
		ec2:              ec2Client,
	}

	ec2Client.On("DescribeInstances", &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{Name: aws.String("private-dns-name"), Values: []*string{aws.String("ip-1.2.3.4.ec2.internal")}},
		},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{InstanceId: aws.String("i-3")},
				},
			},
		},
	}, nil)

	instances := []*elb.Instance{
		{InstanceId: aws.String("i-1")},
		{InstanceId: aws.String("i-2")},
	}

	elbClient.On("DescribeLoadBalancers", &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String("rabbitmq")},
	}).Return(&elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
			{
				Instances: instances,
			},
		},
	}, nil)

	elbClient.On("DeregisterInstancesFromLoadBalancer", &elb.DeregisterInstancesFromLoadBalancerInput{
		LoadBalancerName: aws.String("rabbitmq"),
		Instances:        instances,
	}).Return(&elb.DeregisterInstancesFromLoadBalancerOutput{}, nil)

	elbClient.On("RegisterInstancesWithLoadBalancer", &elb.RegisterInstancesWithLoadBalancerInput{
		LoadBalancerName: aws.String("rabbitmq"),
		Instances: []*elb.Instance{
			{InstanceId: aws.String("i-3")},
		},
	}).Return(&elb.RegisterInstancesWithLoadBalancerOutput{}, nil)

	err := c.SetMaster("rabbit@ip-1.2.3.4.ec2.internal")
	assert.NoError(t, err)

	elbClient.AssertExpectations(t)
	ec2Client.AssertExpectations(t)
}

func TestNodeHostname(t *testing.T) {
	tests := []struct {
		node     string
		hostname string
	}{
		{"rabbit@master", "master"},
		{"rabbit@ip-1.2.3.4.ec2.internal", "ip-1.2.3.4.ec2.internal"},
	}

	for _, tt := range tests {
		hostname := nodeHostname(tt.node)
		assert.Equal(t, tt.hostname, hostname)
	}
}

type mockEC2Client struct {
	mock.Mock
}

func (m *mockEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
}

type mockELBClient struct {
	mock.Mock
}

func (m *mockELBClient) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*elb.DescribeLoadBalancersOutput), args.Error(1)
}

func (m *mockELBClient) DeregisterInstancesFromLoadBalancer(input *elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*elb.DeregisterInstancesFromLoadBalancerOutput), args.Error(1)
}

func (m *mockELBClient) RegisterInstancesWithLoadBalancer(input *elb.RegisterInstancesWithLoadBalancerInput) (*elb.RegisterInstancesWithLoadBalancerOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*elb.RegisterInstancesWithLoadBalancerOutput), args.Error(1)
}
