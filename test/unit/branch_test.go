package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

type BranchNodeTestSuite struct {
	suite.Suite
	wm         *workflow.WorkflowManager
	branchNode *workflow.BranchNode
	mockNode1  *MockNode
	mockNode2  *MockNode
	ctx        context.Context
}

func TestBranchNodeTestSuite(t *testing.T) {
	suite.Run(t, new(BranchNodeTestSuite))
}

func (suite *BranchNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()
	suite.mockNode1 = &MockNode{}
	suite.mockNode2 = &MockNode{}
	suite.ctx = context.Background()

	suite.mockNode1.On("GetID").Return("node1")
	suite.mockNode2.On("GetID").Return("node2")

	suite.branchNode = &workflow.BranchNode{
		Node: workflow.Node[interface{}]{
			ID:   "branch",
			Type: workflow.Branch,
		},
		Branches: []workflow.NodeInterface{suite.mockNode1, suite.mockNode2},
	}
}

func (suite *BranchNodeTestSuite) givenNodesAreSetUp() {
	suite.mockNode1.AssertNotCalled(suite.T(), "Execute", suite.ctx, suite.wm, "testdata")
	suite.mockNode2.AssertNotCalled(suite.T(), "Execute", suite.ctx, suite.wm, "testdata")
}

func (suite *BranchNodeTestSuite) givenNode1Fails() {
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("node1 error"))
}

func (suite *BranchNodeTestSuite) whenBranchNodeIsExecuted() {
	result, err := suite.branchNode.Execute(suite.ctx, suite.wm, "testdata")
	suite.NoError(err)
	suite.NotNil(result)

	// El BranchNode retorna un map con estructura específica
	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map[string]interface{}")

	// Verificar que contiene los resultados esperados
	if results, exists := resultMap["results"]; exists {
		resultsArray, ok := results.([]interface{})
		suite.True(ok, "Results should be an array")
		suite.ElementsMatch([]interface{}{"result1", "result2"}, resultsArray)
	}
}

func (suite *BranchNodeTestSuite) whenBranchNodeIsExecutedWithError() {
	result, err := suite.branchNode.Execute(suite.ctx, suite.wm, "testdata")
	suite.Error(err)
	suite.Nil(result)
}

func (suite *BranchNodeTestSuite) thenBothNodesShouldBeExecuted() {
	// Los nodos son ejecutados con contextos derivados (WithCancel), no el contexto original
	suite.mockNode1.AssertCalled(suite.T(), "Execute", mock.AnythingOfType("*context.cancelCtx"), suite.wm, "testdata")
	suite.mockNode2.AssertCalled(suite.T(), "Execute", mock.AnythingOfType("*context.cancelCtx"), suite.wm, "testdata")
}

func (suite *BranchNodeTestSuite) thenNode1ShouldFail() {
	// El nodo es ejecutado con contexto derivado (WithCancel), no el contexto original
	suite.mockNode1.AssertCalled(suite.T(), "Execute", mock.AnythingOfType("*context.cancelCtx"), suite.wm, "testdata")
}

func (suite *BranchNodeTestSuite) TestBranchNodeExecutesBothNodes() {
	suite.givenNodesAreSetUp()
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return("result1", nil)
	suite.mockNode2.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return("result2", nil)
	suite.whenBranchNodeIsExecuted()
	suite.thenBothNodesShouldBeExecuted()
}

func (suite *BranchNodeTestSuite) TestBranchNodeHandlesError() {
	suite.givenNode1Fails()
	suite.mockNode2.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return("result2", nil)
	suite.whenBranchNodeIsExecutedWithError()
	suite.thenNode1ShouldFail()
}
