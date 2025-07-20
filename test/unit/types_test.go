package unit

import (
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

	expectedMessage := "workflow error in node node-1 (task): test message"
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
	suite.Nil(ctx.EndTime)

	// Marcar como terminado
	endTime := time.Now()
	ctx.EndTime = &endTime
	suite.NotNil(ctx.EndTime)
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
		Min:      nil,
		Max:      nil,
		Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
	}

	suite.Equal("email", rule.Field)
	suite.Equal("email", rule.Type)
	suite.True(rule.Required)
	suite.Nil(rule.Min)
	suite.Nil(rule.Max)
	suite.NotEmpty(rule.Pattern)
}

// TestTaskFuncSignature verifica que TaskFunc tenga la firma correcta
func (suite *TypesTestSuite) TestTaskFuncSignature() {
	var taskFunc workflow.TaskFunc = func(data map[string]interface{}) (map[string]interface{}, error) {
		data["processed"] = true
		return data, nil
	}

	// Probar la función
	input := map[string]interface{}{"test": "value"}
	result, err := taskFunc(input)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(true, result["processed"])
	suite.Equal("value", result["test"])
}

// TestHookFuncSignatures verifica las firmas de las funciones de hook
func (suite *TypesTestSuite) TestHookFuncSignatures() {
	// HookFunc simple
	var hookFunc workflow.HookFunc = func() error {
		return nil
	}
	suite.NoError(hookFunc())

	// HookFuncWithData
	var hookFuncWithData workflow.HookFuncWithData = func(data interface{}) error {
		suite.NotNil(data)
		return nil
	}
	suite.NoError(hookFuncWithData(map[string]interface{}{"key": "value"}))

	// HookFuncWithContext
	var hookFuncWithContext workflow.HookFuncWithContext = func(ctx *workflow.HookContext) error {
		suite.NotNil(ctx)
		return nil
	}

	hookContext := &workflow.HookContext{
		NodeID:   "test-node",
		NodeType: workflow.Task,
		HookType: workflow.BeforeExecution,
	}
	suite.NoError(hookFuncWithContext(hookContext))
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
