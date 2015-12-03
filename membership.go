package clusterctl

// DefaultMembershipController is a membership controller that uses the
// rabbitmqctl command.
var DefaultMembershipController = &RabbitmqCtlMembershipController{
	rabbitmqctl: rabbitmqctl,
}

type JoinNodeOptions struct {
	Node       string
	MasterNode string
}

type RemoveNodeOptions struct {
	Node       string
	MasterNode string
}

// MembershipController is an interface for handling cluster membership of
// individual rabbitmq nodes.
type MembershipController interface {
	JoinNode(JoinNodeOptions) error
	RemoveNode(RemoveNodeOptions) error
}

// membershipController is a MembershipController implementation that uses the
// rabbitmqctl command to add and remove nodes.
type RabbitmqCtlMembershipController struct {
	// function to execute to invoke rabbitmqctl.
	rabbitmqctl rabbitmqctlFunc
}

// JoinNode joins the node to the cluster.
func (c *RabbitmqCtlMembershipController) JoinNode(options JoinNodeOptions) error {
	if err := c.rabbitmqctl(options.Node, "stop_app"); err != nil {
		return err
	}

	if err := c.rabbitmqctl(options.Node, "join_cluster", options.MasterNode); err != nil {
		return err
	}

	if err := c.rabbitmqctl(options.Node, "start_app"); err != nil {
		return err
	}

	return nil
}

// RemoveNode removes the node from the cluster.
func (c *RabbitmqCtlMembershipController) RemoveNode(options RemoveNodeOptions) error {
	if err := c.rabbitmqctl(options.Node, "stop_app"); err != nil {
		return err
	}

	if err := c.rabbitmqctl(options.MasterNode, "forget_cluster_node", options.Node); err != nil {
		return err
	}

	if err := c.rabbitmqctl(options.Node, "reset"); err != nil {
		return err
	}

	return nil
}
