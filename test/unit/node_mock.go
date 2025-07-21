package unit

import (
	"context"
	"github.com/stretchr/testify/mock"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
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

func (m *MockNode) Execute(ctx context.Context, wm *workflow.WorkflowManager, data interface{}) (interface{}, error) {
	args := m.Called(ctx, wm, data)
	return args.Get(0), args.Error(1)
}
