package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

// TypesTestSuite agrupa todas las pruebas del sistema de tipos
type TypesTestSuite struct {
	suite.Suite
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

// TestNodeTypeString verifica que NodeType implementa String correctamente
func (suite *TypesTestSuite) TestNodeTypeString() {
	testCases := []struct {
		nodeType workflow.NodeType
		expected string
	}{
		{workflow.Task, "task"},
		{workflow.SubDag, "subdag"},
		{workflow.Conditional, "conditional"},
		{workflow.Foreach, "foreach"},
		{workflow.Branch, "branch"},
		{workflow.Parallel, "parallel"},
		{workflow.HTTP, "http"},
		{workflow.Delay, "delay"},
		{workflow.Transform, "transform"},
		{workflow.Validation, "validation"},
		{workflow.Merge, "merge"},
		{workflow.Split, "split"},
		{workflow.Filter, "filter"},
	}

	for _, tc := range testCases {
		suite.Equal(tc.expected, tc.nodeType.String())
	}
}

// TestNodeTypeIsValid verifica la validación de tipos de nodos
func (suite *TypesTestSuite) TestNodeTypeIsValid() {
	// Tipos válidos
	validTypes := []workflow.NodeType{
		workflow.Task,
		workflow.SubDag,
		workflow.Conditional,
		workflow.Foreach,
		workflow.Branch,
		workflow.Parallel,
		workflow.HTTP,
		workflow.Delay,
		workflow.Transform,
		workflow.Validation,
		workflow.Merge,
		workflow.Split,
		workflow.Filter,
	}

	for _, nodeType := range validTypes {
		suite.True(nodeType.IsValid(), "NodeType %s should be valid", nodeType)
	}

	// Tipo inválido
	invalidType := workflow.NodeType(999)
	suite.False(invalidType.IsValid(), "Invalid NodeType should return false")
}

// TestWorkflowErrorCreation verifica la creación de errores de workflow
func (suite *TypesTestSuite) TestWorkflowErrorCreation() {
	originalError := errors.New("original error")
	nodeID := "test-node"
	nodeType := workflow.Task
	message := "test error message"

	workflowErr := workflow.NewWorkflowError(nodeID, nodeType, message, originalError)

	// Verificar campos básicos
	suite.Equal(nodeID, workflowErr.NodeID)
	suite.Equal(nodeType, workflowErr.NodeType)
	suite.Equal(message, workflowErr.Message)
	suite.Equal(originalError, workflowErr.Cause)
	suite.NotNil(workflowErr.Context)
	suite.NotEmpty(workflowErr.Stack)
	suite.False(workflowErr.Timestamp.IsZero())
}

// TestWorkflowErrorInterface verifica que WorkflowError implementa la interfaz error
func (suite *TypesTestSuite) TestWorkflowErrorInterface() {
	workflowErr := workflow.NewWorkflowError("node-1", workflow.Task, "test message", nil)

	// Verificar que implementa error
	var err error = workflowErr
	suite.NotNil(err)

	// Actualizar la expectativa para que coincida con el formato actual
	expectedMessage := "[task] test message (node: node-1)"
	suite.Equal(expectedMessage, workflowErr.Error())
}

// TestWorkflowErrorUnwrap verifica la funcionalidad Unwrap
func (suite *TypesTestSuite) TestWorkflowErrorUnwrap() {
	originalError := errors.New("original error")
	workflowErr := workflow.NewWorkflowError("node-1", workflow.Task, "test message", originalError)

	// Verificar Unwrap
	suite.Equal(originalError, workflowErr.Unwrap())

	// Verificar errors.Is
	suite.True(errors.Is(workflowErr, originalError))
}

// TestWorkflowErrorWithContext verifica la funcionalidad WithContext
func (suite *TypesTestSuite) TestWorkflowErrorWithContext() {
	workflowErr := workflow.NewWorkflowError("node-1", workflow.Task, "test message", nil)

	// Añadir contexto
	workflowErr.WithContext("user_id", "123")
	workflowErr.WithContext("retry_count", 3)

	// Verificar contexto
	suite.Equal("123", workflowErr.Context["user_id"])
	suite.Equal(3, workflowErr.Context["retry_count"])
}

// TestExecutionContext verifica la estructura ExecutionContext
func (suite *TypesTestSuite) TestExecutionContext() {
	ctx := &workflow.ExecutionContext{
		WorkflowID:  "workflow-1",
		NodeID:      "node-1",
		ExecutionID: "exec-1",
		Data:        map[string]interface{}{"key": "value"},
		Metadata:    map[string]interface{}{"version": "1.0"},
		StartTime:   time.Now(),
	}

	suite.Equal("workflow-1", ctx.WorkflowID)
	suite.Equal("node-1", ctx.NodeID)
	suite.Equal("exec-1", ctx.ExecutionID)
	suite.NotNil(ctx.Data)
	suite.NotNil(ctx.Metadata)
	suite.False(ctx.StartTime.IsZero())
	// EndTime debe ser el valor cero de time.Time inicialmente
	suite.True(ctx.EndTime.IsZero())

	// Marcar como terminado
	endTime := time.Now()
	ctx.EndTime = endTime
	suite.False(ctx.EndTime.IsZero())
}

// TestHookContext verifica la estructura HookContext
func (suite *TypesTestSuite) TestHookContext() {
	hookCtx := &workflow.HookContext{
		NodeID:      "node-1",
		NodeType:    workflow.Task,
		HookType:    workflow.BeforeExecution,
		Data:        map[string]interface{}{"key": "value"},
		Metadata:    map[string]interface{}{"version": "1.0"},
		ExecutionID: "exec-1",
		Timestamp:   time.Now(),
	}

	suite.Equal("node-1", hookCtx.NodeID)
	suite.Equal(workflow.Task, hookCtx.NodeType)
	suite.Equal(workflow.BeforeExecution, hookCtx.HookType)
	suite.NotNil(hookCtx.Data)
	suite.NotNil(hookCtx.Metadata)
	suite.Equal("exec-1", hookCtx.ExecutionID)
	suite.False(hookCtx.Timestamp.IsZero())
}

// TestValidationRule verifica la estructura ValidationRule
func (suite *TypesTestSuite) TestValidationRule() {
	rule := &workflow.ValidationRule{
		Field:    "email",
		Type:     "email",
		Required: true,
		Min:      0, // Cambiar nil por 0 ya que es int
		Max:      0, // Cambiar nil por 0 ya que es int
		MinValue: nil,
		MaxValue: nil,
		Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
	}

	suite.Equal("email", rule.Field)
	suite.Equal("email", rule.Type)
	suite.True(rule.Required)
	suite.Equal(0, rule.Min)
	suite.Equal(0, rule.Max)
	suite.NotEmpty(rule.Pattern)
}

// TestTaskFuncSignature verifica que TaskFunc tenga la firma correcta
func (suite *TypesTestSuite) TestTaskFuncSignature() {
	// Corregir la signatura para que coincida con TaskFunc (context.Context, interface{}) (interface{}, error)
	var taskFunc workflow.TaskFunc = func(ctx context.Context, data interface{}) (interface{}, error) {
		// Convertir a map si es posible
		if dataMap, ok := data.(map[string]interface{}); ok {
			dataMap["processed"] = true
			return dataMap, nil
		}
		// Si no es un map, crear uno nuevo
		result := map[string]interface{}{
			"original":  data,
			"processed": true,
		}
		return result, nil
	}

	// Probar la función con context
	input := map[string]interface{}{"test": "value"}
	result, err := taskFunc(context.Background(), input)

	suite.NoError(err)
	suite.NotNil(result)
	if resultMap, ok := result.(map[string]interface{}); ok {
		suite.Equal(true, resultMap["processed"])
		suite.Equal("value", resultMap["test"])
	}
}

// TestHookFuncSignatures verifica las firmas de las funciones de hook
func (suite *TypesTestSuite) TestHookFuncSignatures() {
	// HookFunc con signatura correcta (HookContext) error
	var hookFunc workflow.HookFunc = func(ctx workflow.HookContext) error {
		return nil
	}

	hookContext := workflow.HookContext{
		NodeID:   "test-node",
		NodeType: workflow.Task,
		HookType: workflow.BeforeExecution,
	}
	suite.NoError(hookFunc(hookContext))

	// Comentar las funciones que no existen en el sistema actual
	// Estas se pueden implementar en el futuro si son necesarias
}

// TestAliasNodes verifica que los alias de nodos funcionen correctamente
func (suite *TypesTestSuite) TestAliasNodes() {
	// Verificar que los alias apunten a los valores correctos
	suite.Equal(workflow.Branch, workflow.Decision)
	suite.Equal(workflow.Foreach, workflow.Loop)
	suite.Equal(workflow.SubDag, workflow.Subflow)

	// Verificar que los alias tengan el string correcto
	suite.Equal("branch", workflow.Decision.String())
	suite.Equal("foreach", workflow.Loop.String())
	suite.Equal("subdag", workflow.Subflow.String())
}
