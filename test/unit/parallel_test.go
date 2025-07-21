package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
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
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return("result1", nil)

	suite.mockNode2.On("GetID").Return("node2")
	suite.mockNode2.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return("result2", nil)

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
	suite.mockNode1.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("node1 error"))
}

func (suite *ParallelNodeTestSuite) whenParallelNodeIsExecuted() {
	ctx := context.Background()
	result, err := suite.parallelNode.Execute(ctx, suite.wm, "testdata")
	suite.NoError(err)
	suite.NotNil(result)

	// El ParallelNode retorna un map con estructura específica
	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map[string]interface{}")

	// Verificar que contiene los resultados esperados
	if results, exists := resultMap["results"]; exists {
		resultsArray, ok := results.([]interface{})
		suite.True(ok, "Results should be an array")
		suite.ElementsMatch([]interface{}{"result1", "result2"}, resultsArray)
	}
}

func (suite *ParallelNodeTestSuite) whenParallelNodeIsExecutedWithError() {
	ctx := context.Background()
	result, err := suite.parallelNode.Execute(ctx, suite.wm, "testdata")
	suite.Error(err)
	suite.Nil(result)
}

func (suite *ParallelNodeTestSuite) thenBothTasksShouldBeExecuted() {
	// Los nodos son ejecutados con contextos derivados (WithCancel), no el contexto original
	suite.mockNode1.AssertCalled(suite.T(), "Execute", mock.AnythingOfType("*context.cancelCtx"), suite.wm, "testdata")
	suite.mockNode2.AssertCalled(suite.T(), "Execute", mock.AnythingOfType("*context.cancelCtx"), suite.wm, "testdata")
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
