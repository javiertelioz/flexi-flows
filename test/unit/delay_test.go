package unit

import (
	"context"
	"testing"
	"time"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type DelayNodeTestSuite struct {
	suite.Suite
	wm        *workflow.WorkflowManager
	delayNode *workflow.DelayNode
	ctx       context.Context
}

func TestDelayNodeTestSuite(t *testing.T) {
	suite.Run(t, new(DelayNodeTestSuite))
}

func (suite *DelayNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()
	suite.ctx = context.Background()

	delayType, err := workflow.ParseNodeType("delay")
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.delayNode = &workflow.DelayNode{
		Node: workflow.Node[interface{}]{
			ID:   "delay_test",
			Type: delayType,
		},
		Duration: 100 * time.Millisecond, // Short delay for testing
	}
}

func (suite *DelayNodeTestSuite) TestDelayNodeExecuteSuccess() {
	start := time.Now()

	result, err := suite.delayNode.Execute(suite.ctx, suite.wm, "test_data")

	elapsed := time.Since(start)

	suite.NoError(err)
	suite.NotNil(result)

	// Verificar que el delay se ejecutó (al menos 90ms para dar margen)
	suite.GreaterOrEqual(elapsed, 90*time.Millisecond)

	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map")

	suite.Equal("delay completed successfully", resultMap["message"])
	suite.Equal("100ms", resultMap["duration"])
	suite.Equal("test_data", resultMap["data"])
}

func (suite *DelayNodeTestSuite) TestDelayNodeExecuteContextCancellation() {
	// Crear contexto que se cancela después de 50ms
	cancelCtx, cancel := context.WithTimeout(suite.ctx, 50*time.Millisecond)
	defer cancel()

	// Usar delay más largo que el timeout
	suite.delayNode.Duration = 200 * time.Millisecond

	start := time.Now()
	result, err := suite.delayNode.Execute(cancelCtx, suite.wm, "test_data")
	elapsed := time.Since(start)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "context cancelled")

	// Verificar que se canceló antes de completar el delay
	suite.Less(elapsed, 150*time.Millisecond)
}

func (suite *DelayNodeTestSuite) TestDelayNodeExecuteZeroDuration() {
	suite.delayNode.Duration = 0

	start := time.Now()
	result, err := suite.delayNode.Execute(suite.ctx, suite.wm, "test_data")
	elapsed := time.Since(start)

	suite.NoError(err)
	suite.NotNil(result)

	// Con duración 0, debería completarse inmediatamente
	suite.Less(elapsed, 10*time.Millisecond)
}

func (suite *DelayNodeTestSuite) TestDelayNodeExecuteContextCancelledBeforeStart() {
	// Crear contexto ya cancelado
	cancelCtx, cancel := context.WithCancel(suite.ctx)
	cancel()

	result, err := suite.delayNode.Execute(cancelCtx, suite.wm, "test_data")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "context cancelled before delay")
}
