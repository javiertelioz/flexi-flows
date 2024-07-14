package unit

import (
	"errors"
	"testing"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type WorkflowManagerTestSuite struct {
	suite.Suite
	wm        *workflow.WorkflowManager
	mockNode1 *MockNode
	mockNode2 *MockNode
	mockEdge  *workflow.Edge
}

func TestWorkflowManagerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowManagerTestSuite))
}

func (suite *WorkflowManagerTestSuite) SetupTest() {
	suite.mockNode1 = new(MockNode)
	suite.mockNode2 = new(MockNode)

	suite.mockNode1.On("GetID").Return("node1")
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything).Return("result1", nil)

	suite.mockNode2.On("GetID").Return("node2")
	suite.mockNode2.On("Execute", mock.Anything, mock.Anything).Return("result2", nil)

	suite.mockEdge = &workflow.Edge{
		From: suite.mockNode1,
		To:   suite.mockNode2,
	}

	suite.wm = workflow.NewWorkflowManager()
	suite.wm.AddNode(suite.mockNode1)
	suite.wm.AddNode(suite.mockNode2)
	suite.wm.AddEdge(suite.mockEdge)
}

func (suite *WorkflowManagerTestSuite) givenNodesAreSetUp() {
	suite.mockNode1.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
}

func (suite *WorkflowManagerTestSuite) givenNode1Fails() {
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("node1 error"))
}

func (suite *WorkflowManagerTestSuite) givenStartNodeIsNil() {
	suite.wm = workflow.NewWorkflowManager()
}

func (suite *WorkflowManagerTestSuite) givenNodeIsNil() {
	suite.mockNode1 = nil
}

func (suite *WorkflowManagerTestSuite) whenWorkflowIsExecuted() {
	err := suite.wm.Execute("node1", "testdata")
	suite.NoError(err)
}

func (suite *WorkflowManagerTestSuite) whenWorkflowIsExecutedWithError() {
	err := suite.wm.Execute("node1", "testdata")
	suite.Error(err)
}

func (suite *WorkflowManagerTestSuite) whenWorkflowIsExecutedWithNilStartNode() {
	err := suite.wm.Execute("node1", "testdata")
	suite.EqualError(err, "start node not found")
}

func (suite *WorkflowManagerTestSuite) whenNodeIsExecutedWithNilNode() {
	result, err := suite.wm.ExecuteNode(nil, "testdata")
	suite.EqualError(err, "node is nil")
	suite.Nil(result)
}

func (suite *WorkflowManagerTestSuite) thenBothNodesShouldBeExecuted() {
	suite.mockNode1.AssertCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertCalled(suite.T(), "Execute", suite.wm, "result1")
}

func (suite *WorkflowManagerTestSuite) thenNode1ShouldFail() {
	suite.mockNode1.AssertCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
}

func (suite *WorkflowManagerTestSuite) TestWorkflowExecution() {
	suite.givenNodesAreSetUp()
	suite.whenWorkflowIsExecuted()
	suite.thenBothNodesShouldBeExecuted()
}

/*func (suite *WorkflowManagerTestSuite) TestWorkflowExecutionWithError() {
	suite.givenNodesAreSetUp()
	suite.givenNode1Fails()
	suite.whenWorkflowIsExecutedWithError()
	suite.thenNode1ShouldFail()
}*/

func (suite *WorkflowManagerTestSuite) TestWorkflowExecutionWithNilStartNode() {
	suite.givenStartNodeIsNil()
	suite.whenWorkflowIsExecutedWithNilStartNode()
}

func (suite *WorkflowManagerTestSuite) TestNodeExecutionWithNilNode() {
	suite.givenNodeIsNil()
	suite.whenNodeIsExecutedWithNilNode()
}
