package unit

import (
	"errors"
	"github.com/javiertelioz/go-flows/pkg/workflow/nodes"
	"testing"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BranchNodeTestSuite struct {
	suite.Suite
	wm         *workflow.WorkflowManager
	mockNode1  *MockNode
	mockNode2  *MockNode
	branchNode *nodes.BranchNode
}

func TestBranchNodeTestSuite(t *testing.T) {
	suite.Run(t, new(BranchNodeTestSuite))
}

func (suite *BranchNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()

	suite.mockNode1 = new(MockNode)
	suite.mockNode2 = new(MockNode)

	suite.mockNode1.On("GetID").Return("node1")
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything).Return(nil, nil)

	suite.mockNode2.On("GetID").Return("node2")
	suite.mockNode2.On("Execute", mock.Anything, mock.Anything).Return(nil, nil)

	suite.branchNode = &nodes.BranchNode{
		Node: workflow.Node[interface{}]{
			ID:   "branch",
			Type: workflow.Branch,
		},
		Branches: []workflow.NodeInterface{suite.mockNode1, suite.mockNode2},
	}
}

func (suite *BranchNodeTestSuite) givenNodesAreSetUp() {
	suite.mockNode1.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
}

func (suite *BranchNodeTestSuite) givenNode1Fails() {
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("node1 error"))
}

func (suite *BranchNodeTestSuite) whenBranchNodeIsExecuted() {
	result, err := suite.branchNode.Execute(suite.wm, "testdata")
	suite.NoError(err)
	suite.Nil(result)
}

func (suite *BranchNodeTestSuite) whenBranchNodeIsExecutedWithError() {
	result, err := suite.branchNode.Execute(suite.wm, "testdata")
	suite.Error(err)
	suite.Nil(result)
}

func (suite *BranchNodeTestSuite) thenBothNodesShouldBeExecuted() {
	suite.mockNode1.AssertCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertCalled(suite.T(), "Execute", suite.wm, "testdata")
}

func (suite *BranchNodeTestSuite) thenNode1ShouldFail() {
	suite.mockNode1.AssertCalled(suite.T(), "Execute", suite.wm, "testdata")
	suite.mockNode2.AssertNotCalled(suite.T(), "Execute", suite.wm, "testdata")
}

func (suite *BranchNodeTestSuite) TestBranchNodeExecution() {
	suite.givenNodesAreSetUp()
	suite.whenBranchNodeIsExecuted()
	suite.thenBothNodesShouldBeExecuted()
}

/*func (suite *BranchNodeTestSuite) TestBranchNodeExecutionWithError() {
	suite.givenNodesAreSetUp()
	suite.givenNode1Fails()
	suite.whenBranchNodeIsExecutedWithError()
	suite.thenNode1ShouldFail()
}*/
