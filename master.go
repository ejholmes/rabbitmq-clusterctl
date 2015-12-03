package clusterctl

import "errors"

// NullMasterController is a MasterController that does nothing.
var NullMasterController = &staticMasterController{
	node: "rabbit@localhost",
}

// MasterController is an interface for setting the rabbitmq master node.
type MasterController interface {
	Master() (node string, err error)
	SetMaster(node string) error
}

// ELBMasterController implements the MasterController interface using an ELB as
// the source of truth.
type ELBMasterController struct {
}

// Master returns the node name of the current master.
func (c *ELBMasterController) Master() (string, error) {
	return "", nil
}

// SetMaster sets the node to be the new master.
func (c *ELBMasterController) SetMaster(node string) error {
	return nil
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
