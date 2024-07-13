package test

import (
	"github.com/stretchr/testify/mock"

	"github.com/javiertelioz/workflows/pkg/workflow"
)

type MockNode struct {
	mock.Mock
}

func (m *MockNode) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockNode) GetType() workflow.NodeType {
	return workflow.Task
}

func (m *MockNode) Execute(wm *workflow.WorkflowManager, data interface{}) (interface{}, error) {
	args := m.Called(wm, data)
	return args.Get(0), args.Error(1)
}
