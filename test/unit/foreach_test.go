package unit

import (
	"errors"
	"github.com/javiertelioz/go-flows/pkg/workflow/nodes"
	"testing"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type ForeachNodeTestSuite struct {
	suite.Suite
	wm          *workflow.WorkflowManager
	foreachNode *nodes.ForeachNode
}

func TestForeachNodeTestSuite(t *testing.T) {
	suite.Run(t, new(ForeachNodeTestSuite))
}

func (suite *ForeachNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()
	items := []interface{}{1, 2, 3}
	iterateFunc := func(item interface{}) (interface{}, error) {
		return item.(int) * 2, nil
	}

	suite.foreachNode = &nodes.ForeachNode{
		Node: workflow.Node[interface{}]{
			ID:   "foreach",
			Type: workflow.Foreach,
		},
		Collection:  items,
		IterateFunc: iterateFunc,
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
	result, err := suite.foreachNode.Execute(suite.wm, nil)
	suite.NoError(err)
	suite.Nil(result)
}

func (suite *ForeachNodeTestSuite) whenForeachNodeIsExecutedWithError() {
	result, err := suite.foreachNode.Execute(suite.wm, nil)
	suite.Error(err)
	suite.Nil(result)
}

func (suite *ForeachNodeTestSuite) thenAllItemsShouldBeProcessed() {
	// No additional assertions needed here, as they are in the `when` step
}

func (suite *ForeachNodeTestSuite) thenExecutionShouldFail() {
	// No additional assertions needed here, as they are in the `when` step
}

func (suite *ForeachNodeTestSuite) TestForeachNodeExecution() {
	suite.givenForeachNodeIsSetUp()
	suite.whenForeachNodeIsExecuted()
	suite.thenAllItemsShouldBeProcessed()
}

func (suite *ForeachNodeTestSuite) TestForeachNodeExecutionWithError() {
	suite.givenForeachNodeIsSetUp()
	suite.givenIterateFuncFailsForItem()
	suite.whenForeachNodeIsExecutedWithError()
	suite.thenExecutionShouldFail()
}
