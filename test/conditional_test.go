package test

import (
	"errors"
	"testing"

	"github.com/javiertelioz/workflows/pkg/workflow"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ConditionalNodeTestSuite struct {
	suite.Suite
	wm              *workflow.WorkflowManager
	mockTrueNode    *MockNode
	mockFalseNode   *MockNode
	conditionalNode *workflow.ConditionalNode
}

func TestConditionalNodeTestSuite(t *testing.T) {
	suite.Run(t, new(ConditionalNodeTestSuite))
}

func (suite *ConditionalNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager(nil, nil, nil)

	suite.mockTrueNode = new(MockNode)
	suite.mockFalseNode = new(MockNode)

	suite.mockTrueNode.On("GetID").Return("trueNode")
	suite.mockTrueNode.On("Execute", mock.Anything, mock.Anything).Return("trueResult", nil)

	suite.mockFalseNode.On("GetID").Return("falseNode")
	suite.mockFalseNode.On("Execute", mock.Anything, mock.Anything).Return("falseResult", nil)

	condition := func(data interface{}) bool {
		return data.(bool)
	}

	suite.conditionalNode = &workflow.ConditionalNode{
		Node: workflow.Node[interface{}]{
			ID:   "conditional",
			Type: workflow.Conditional,
		},
		Condition: condition,
		TrueNext:  suite.mockTrueNode,
		FalseNext: suite.mockFalseNode,
	}
}

func (suite *ConditionalNodeTestSuite) givenConditionIsTrue() {
	suite.mockTrueNode.AssertNotCalled(suite.T(), "Execute", suite.wm, true)
	suite.mockFalseNode.AssertNotCalled(suite.T(), "Execute", suite.wm, true)
}

func (suite *ConditionalNodeTestSuite) givenConditionIsFalse() {
	suite.mockTrueNode.AssertNotCalled(suite.T(), "Execute", suite.wm, false)
	suite.mockFalseNode.AssertNotCalled(suite.T(), "Execute", suite.wm, false)
}

func (suite *ConditionalNodeTestSuite) givenTrueNodeFails() {
	suite.mockTrueNode.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("trueNode error"))
}

func (suite *ConditionalNodeTestSuite) whenConditionalNodeIsExecutedWithTrue() {
	result, err := suite.conditionalNode.Execute(suite.wm, true)
	suite.NoError(err)
	suite.Equal("trueResult", result)
}

func (suite *ConditionalNodeTestSuite) whenConditionalNodeIsExecutedWithFalse() {
	result, err := suite.conditionalNode.Execute(suite.wm, false)
	suite.NoError(err)
	suite.Equal("falseResult", result)
}

func (suite *ConditionalNodeTestSuite) whenConditionalNodeIsExecutedWithTrueAndFails() {
	result, err := suite.conditionalNode.Execute(suite.wm, true)
	suite.Error(err)
	suite.Nil(result)
}

func (suite *ConditionalNodeTestSuite) thenTrueNodeShouldBeExecuted() {
	suite.mockTrueNode.AssertCalled(suite.T(), "Execute", suite.wm, true)
	suite.mockFalseNode.AssertNotCalled(suite.T(), "Execute", suite.wm, true)
}

func (suite *ConditionalNodeTestSuite) thenFalseNodeShouldBeExecuted() {
	suite.mockTrueNode.AssertNotCalled(suite.T(), "Execute", suite.wm, false)
	suite.mockFalseNode.AssertCalled(suite.T(), "Execute", suite.wm, false)
}

func (suite *ConditionalNodeTestSuite) thenTrueNodeShouldFail() {
	suite.mockTrueNode.AssertCalled(suite.T(), "Execute", suite.wm, true)
	suite.mockFalseNode.AssertNotCalled(suite.T(), "Execute", suite.wm, true)
}

func (suite *ConditionalNodeTestSuite) TestConditionalNodeExecutionTrue() {
	suite.givenConditionIsTrue()
	suite.whenConditionalNodeIsExecutedWithTrue()
	suite.thenTrueNodeShouldBeExecuted()
}

func (suite *ConditionalNodeTestSuite) TestConditionalNodeExecutionFalse() {
	suite.givenConditionIsFalse()
	suite.whenConditionalNodeIsExecutedWithFalse()
	suite.thenFalseNodeShouldBeExecuted()
}

/*func (suite *ConditionalNodeTestSuite) TestConditionalNodeExecutionTrueWithError() {
	suite.givenConditionIsTrue()
	suite.givenTrueNodeFails()
	suite.whenConditionalNodeIsExecutedWithTrueAndFails()
	suite.thenTrueNodeShouldFail()
}*/
