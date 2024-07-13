package test

import (
	"errors"
	"testing"

	"github.com/javiertelioz/workflows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type NodeTestSuite struct {
	suite.Suite
	wm   *workflow.WorkflowManager
	node *workflow.Node[int]
}

func TestNodeTestSuite(t *testing.T) {
	suite.Run(t, new(NodeTestSuite))
}

func (suite *NodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager(nil, nil, nil)

	taskFunc := func(data int) (int, error) {
		return data * 2, nil
	}

	beforeFunc := func(data int) (int, error) {
		return data + 1, nil
	}

	afterFunc := func(data int) (int, error) {
		return data - 1, nil
	}

	suite.node = &workflow.Node[int]{
		ID:            "node",
		Type:          workflow.Task,
		TaskFunc:      taskFunc,
		BeforeExecute: beforeFunc,
		AfterExecute:  afterFunc,
	}
}

func (suite *NodeTestSuite) givenNodeIsSetUp() {
	suite.NotNil(suite.node)
}

func (suite *NodeTestSuite) givenTaskFuncFails() {
	suite.node.TaskFunc = func(data int) (int, error) {
		return 0, errors.New("task error")
	}
}

func (suite *NodeTestSuite) whenNodeIsExecuted() {
	result, err := suite.node.Execute(suite.wm, 1)
	suite.NoError(err)
	suite.Equal(3, result)
}

func (suite *NodeTestSuite) whenNodeIsExecutedWithError() {
	result, err := suite.node.Execute(suite.wm, 1)
	suite.Error(err)
	suite.Equal(0, result)
}

func (suite *NodeTestSuite) thenNodeShouldReturnExpectedResult() {
	// No additional assertions needed here, as they are in the `when` step
}

func (suite *NodeTestSuite) thenNodeShouldFail() {
	// No additional assertions needed here, as they are in the `when` step
}

func (suite *NodeTestSuite) TestNodeExecution() {
	suite.givenNodeIsSetUp()
	suite.whenNodeIsExecuted()
	suite.thenNodeShouldReturnExpectedResult()
}

/*func (suite *NodeTestSuite) TestNodeExecutionWithError() {
	suite.givenNodeIsSetUp()
	suite.givenTaskFuncFails()
	suite.whenNodeIsExecutedWithError()
	suite.thenNodeShouldFail()
}*/
