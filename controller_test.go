package clusterctl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestController_Join(t *testing.T) {
	master := new(mockMasterController)
	membership := new(mockMembershipController)
	c := &Controller{
		Node:                 "rabbit@slave",
		MasterController:     master,
		MembershipController: membership,
	}

	master.On("Master").Return("rabbit@master", nil)
	membership.On("JoinNode", JoinNodeOptions{
		Node:       "rabbit@slave",
		MasterNode: "rabbit@master",
	}).Return(nil)

	err := c.Join()
	assert.NoError(t, err)

	master.AssertExpectations(t)
	membership.AssertExpectations(t)
}

func TestController_Remove(t *testing.T) {
	master := new(mockMasterController)
	membership := new(mockMembershipController)
	c := &Controller{
		Node:                 "rabbit@slave",
		MasterController:     master,
		MembershipController: membership,
	}

	master.On("Master").Return("rabbit@master", nil)
	membership.On("RemoveNode", RemoveNodeOptions{
		Node:       "rabbit@slave",
		MasterNode: "rabbit@master",
	}).Return(nil)

	err := c.Remove()
	assert.NoError(t, err)

	master.AssertExpectations(t)
	membership.AssertExpectations(t)
}

func TestController_Promote(t *testing.T) {
	master := new(mockMasterController)
	membership := new(mockMembershipController)
	c := &Controller{
		Node:                 "rabbit@slave",
		MasterController:     master,
		MembershipController: membership,
	}

	master.On("SetMaster", "rabbit@slave").Return(nil)

	err := c.Promote()
	assert.NoError(t, err)
}

type mockMembershipController struct {
	mock.Mock
}

func (m *mockMembershipController) JoinNode(options JoinNodeOptions) error {
	args := m.Called(options)
	return args.Error(0)
}

func (m *mockMembershipController) RemoveNode(options RemoveNodeOptions) error {
	args := m.Called(options)
	return args.Error(0)
}

type mockMasterController struct {
	mock.Mock
}

func (m *mockMasterController) Master() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockMasterController) SetMaster(node string) error {
	args := m.Called(node)
	return args.Error(0)
}
