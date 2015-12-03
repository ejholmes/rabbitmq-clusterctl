package clusterctl

type Controller struct {
	// The current nodes name.
	Node string

	MasterController
	MembershipController
}

// Joins the current node to the cluster.
func (c *Controller) Join() error {
	master, err := c.Master()
	if err != nil {
		return err
	}

	return c.JoinNode(JoinNodeOptions{
		Node:       c.Node,
		MasterNode: master,
	})
}

// Removes the current node from the cluster.
func (c *Controller) Remove() error {
	master, err := c.Master()
	if err != nil {
		return err
	}

	return c.RemoveNode(RemoveNodeOptions{
		Node:       c.Node,
		MasterNode: master,
	})
}

// Promote promotes this node to be the new master.
func (c *Controller) Promote() error {
	return c.SetMaster(c.Node)
}
