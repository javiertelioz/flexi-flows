package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type ForeachNodeTestSuite struct {
	suite.Suite
	wm          *workflow.WorkflowManager
	foreachNode *workflow.ForeachNode
	ctx         context.Context
}

func TestForeachNodeTestSuite(t *testing.T) {
	suite.Run(t, new(ForeachNodeTestSuite))
}

func (suite *ForeachNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()
	suite.ctx = context.Background()

	iterateFunc := func(item interface{}) (interface{}, error) {
		return item.(int) * 2, nil
	}

	suite.foreachNode = &workflow.ForeachNode{
		Node: workflow.Node[interface{}]{
			ID:   "foreach",
			Type: workflow.Foreach,
		},
		IterateFunc: iterateFunc,
		Collection:  []interface{}{1, 2, 3},
	}
}

func (suite *ForeachNodeTestSuite) givenForeachNodeIsSetUp() {
	suite.NotNil(suite.foreachNode)
}

func (suite *ForeachNodeTestSuite) givenIterateFuncFailsForItem() {
	suite.foreachNode.IterateFunc = func(item interface{}) (interface{}, error) {
		if item == 2 {
			return nil, errors.New("error processing item")
		}
		return item.(int) * 2, nil
	}
}

func (suite *ForeachNodeTestSuite) whenForeachNodeIsExecuted() {
	result, err := suite.foreachNode.Execute(suite.ctx, suite.wm, nil)
	suite.NoError(err)
	suite.NotNil(result)
}

func (suite *ForeachNodeTestSuite) whenForeachNodeIsExecutedWithError() {
	result, err := suite.foreachNode.Execute(suite.ctx, suite.wm, nil)
	suite.Error(err)
	suite.Nil(result)
}

func (suite *ForeachNodeTestSuite) TestForeachNodeExecutesSuccessfully() {
	suite.givenForeachNodeIsSetUp()
	suite.whenForeachNodeIsExecuted()
}

func (suite *ForeachNodeTestSuite) TestForeachNodeHandlesError() {
	suite.givenIterateFuncFailsForItem()
	suite.whenForeachNodeIsExecutedWithError()
}
