package unit

import (
	"errors"
	"testing"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ParallelNodeTestSuite struct {
	suite.Suite
	wm           *workflow.WorkflowManager
	mockNode1    *MockNode
	mockNode2    *MockNode
	parallelNode *workflow.ParallelNode
}

func TestParallelNodeTestSuite(t *testing.T) {
	suite.Run(t, new(ParallelNodeTestSuite))
}

func (suite *ParallelNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()

	suite.mockNode1 = new(MockNode)
	suite.mockNode2 = new(MockNode)

	suite.mockNode1.On("GetID").Return("node1")
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything).Return("result1", nil)

	suite.mockNode2.On("GetID").Return("node2")
	suite.mockNode2.On("Execute", mock.Anything, mock.Anything).Return("result2", nil)

	suite.parallelNode = &workflow.ParallelNode{
		Node: workflow.Node[interface{}]{
			ID:   "parallel",
			Type: workflow.Branch,
		},
		ParallelTasks: []workflow.NodeInterface{suite.mockNode1, suite.mockNode2},
	}
}

func (suite *ParallelNodeTestSuite) givenParallelNodeIsSetUp() {
	suite.mockNode1.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
}

func (suite *ParallelNodeTestSuite) givenNode1Fails() {
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("node1 error"))
}

func (suite *ParallelNodeTestSuite) whenParallelNodeIsExecuted() {
	result, err := suite.parallelNode.Execute(suite.wm, "testdata")
	suite.NoError(err)
	suite.ElementsMatch([]interface{}{"result1", "result2"}, result.([]interface{}))
}

func (suite *ParallelNodeTestSuite) whenParallelNodeIsExecutedWithError() {
	result, err := suite.parallelNode.Execute(suite.wm, "testdata")
	suite.Error(err)
	suite.Nil(result)
}

func (suite *ParallelNodeTestSuite) thenBothTasksShouldBeExecuted() {
	suite.mockNode1.AssertCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertCalled(suite.T(), "Execute", suite.wm, "testdata")
}

func (suite *ParallelNodeTestSuite) thenNode1ShouldFail() {
	suite.mockNode1.AssertCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
}

func (suite *ParallelNodeTestSuite) TestParallelNodeExecution() {
	suite.givenParallelNodeIsSetUp()
	suite.whenParallelNodeIsExecuted()
	suite.thenBothTasksShouldBeExecuted()
}

/*func (suite *ParallelNodeTestSuite) TestParallelNodeExecutionWithError() {
	suite.givenParallelNodeIsSetUp()
	suite.givenNode1Fails()
	suite.whenParallelNodeIsExecutedWithError()
	suite.thenNode1ShouldFail()
}*/
