package clusterctl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMembershipController_JoinNode(t *testing.T) {
	m := new(mockRabbitmqCtl)
	c := &RabbitmqCtlMembershipController{
		rabbitmqctl: m.rabbitmqctl,
	}

	m.On("rabbitmqctl", "rabbit@slave", "stop_app", emptyArgs).Return(nil)
	m.On("rabbitmqctl", "rabbit@slave", "join_cluster", []string{"rabbit@master"}).Return(nil)
	m.On("rabbitmqctl", "rabbit@slave", "start_app", emptyArgs).Return(nil)

	err := c.JoinNode(JoinNodeOptions{
		Node:       "rabbit@slave",
		MasterNode: "rabbit@master",
	})
	assert.NoError(t, err)

	m.AssertExpectations(t)
}

func TestMembershipController_RemoveNode(t *testing.T) {
	m := new(mockRabbitmqCtl)
	c := &RabbitmqCtlMembershipController{
		rabbitmqctl: m.rabbitmqctl,
	}

	m.On("rabbitmqctl", "rabbit@slave", "stop_app", emptyArgs).Return(nil)
	m.On("rabbitmqctl", "rabbit@master", "forget_cluster_node", []string{"rabbit@slave"}).Return(nil)
	m.On("rabbitmqctl", "rabbit@slave", "reset", emptyArgs).Return(nil)

	err := c.RemoveNode(RemoveNodeOptions{
		Node:       "rabbit@slave",
		MasterNode: "rabbit@master",
	})
	assert.NoError(t, err)

	m.AssertExpectations(t)
}

// emptyArgs is a niladic []string.
var emptyArgs []string

type mockRabbitmqCtl struct {
	mock.Mock
}

func (m *mockRabbitmqCtl) rabbitmqctl(node string, command string, arg ...string) error {
	args := m.Called(node, command, arg)
	return args.Error(0)
}
