package unit

import (
	"context"
	"testing"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type ValidationNodeTestSuite struct {
	suite.Suite
	wm             *workflow.WorkflowManager
	validationNode *workflow.ValidationNode
	ctx            context.Context
}

func TestValidationNodeTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationNodeTestSuite))
}

func (suite *ValidationNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()
	suite.ctx = context.Background()

	validationType, err := workflow.ParseNodeType("validation")
	if err != nil {
		suite.T().Fatal(err)
	}

	rules := []workflow.ValidationRule{
		{
			Field:    "email",
			Type:     "email",
			Required: true,
			Message:  "Email is required and must be valid",
		},
		{
			Field:    "age",
			Type:     "int",
			Required: true,
			MinValue: 18.0,
			MaxValue: 100.0,
		},
		{
			Field:    "name",
			Type:     "string",
			Required: true,
			MinValue: 2.0,  // min length
			MaxValue: 50.0, // max length
		},
	}

	suite.validationNode = &workflow.ValidationNode{
		Node: workflow.Node[interface{}]{
			ID:   "validation_test",
			Type: validationType,
		},
		Rules:            rules,
		StopOnFirstError: false,
	}
}

func (suite *ValidationNodeTestSuite) TestValidationNodeExecuteAllValid() {
	data := map[string]interface{}{
		"email": "test@example.com",
		"age":   25,
		"name":  "John Doe",
	}

	result, err := suite.validationNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	validationResult, ok := result.(*workflow.ValidationResult)
	suite.True(ok, "Result should be ValidationResult")

	suite.True(validationResult.Valid)
	suite.Empty(validationResult.Errors)
	suite.Equal(data, validationResult.Data)
	suite.Equal(3, validationResult.Summary["total_rules"])
	suite.Equal(0, validationResult.Summary["errors_count"])
}

func (suite *ValidationNodeTestSuite) TestValidationNodeExecuteInvalidEmail() {
	data := map[string]interface{}{
		"email": "invalid-email",
		"age":   25,
		"name":  "John Doe",
	}

	result, err := suite.validationNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	validationResult, ok := result.(*workflow.ValidationResult)
	suite.True(ok, "Result should be ValidationResult")

	suite.False(validationResult.Valid)
	suite.Len(validationResult.Errors, 1)
	suite.Equal("email", validationResult.Errors[0].Field)
	suite.Equal("email", validationResult.Errors[0].Rule)
}

func (suite *ValidationNodeTestSuite) TestValidationNodeExecuteMissingRequiredField() {
	data := map[string]interface{}{
		"age":  25,
		"name": "John Doe",
		// email missing
	}

	result, err := suite.validationNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	validationResult, ok := result.(*workflow.ValidationResult)
	suite.True(ok, "Result should be ValidationResult")

	suite.False(validationResult.Valid)
	suite.Len(validationResult.Errors, 1)
	suite.Equal("email", validationResult.Errors[0].Field)
	suite.Equal("required", validationResult.Errors[0].Rule)
}

func (suite *ValidationNodeTestSuite) TestValidationNodeExecuteAgeOutOfRange() {
	data := map[string]interface{}{
		"email": "test@example.com",
		"age":   15, // Below minimum
		"name":  "John Doe",
	}

	result, err := suite.validationNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	validationResult, ok := result.(*workflow.ValidationResult)
	suite.True(ok, "Result should be ValidationResult")

	suite.False(validationResult.Valid)
	suite.Len(validationResult.Errors, 1)
	suite.Equal("age", validationResult.Errors[0].Field)
	suite.Equal("min_value", validationResult.Errors[0].Rule)
}

func (suite *ValidationNodeTestSuite) TestValidationNodeExecuteMultipleErrors() {
	data := map[string]interface{}{
		"email": "invalid-email",
		"age":   15,  // Below minimum
		"name":  "J", // Too short
	}

	result, err := suite.validationNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	validationResult, ok := result.(*workflow.ValidationResult)
	suite.True(ok, "Result should be ValidationResult")

	suite.False(validationResult.Valid)
	suite.Len(validationResult.Errors, 3) // Should have 3 errors
}

func (suite *ValidationNodeTestSuite) TestValidationNodeExecuteStopOnFirstError() {
	suite.validationNode.StopOnFirstError = true

	data := map[string]interface{}{
		"email": "invalid-email",
		"age":   15,  // Below minimum
		"name":  "J", // Too short
	}

	result, err := suite.validationNode.Execute(suite.ctx, suite.wm, data)

	suite.Error(err) // Should return error when StopOnFirstError is true
	suite.NotNil(result)

	validationResult, ok := result.(*workflow.ValidationResult)
	suite.True(ok, "Result should be ValidationResult")

	suite.False(validationResult.Valid)
}

func (suite *ValidationNodeTestSuite) TestValidationNodeExecuteContextCancellation() {
	// Crear contexto cancelado
	cancelCtx, cancel := context.WithCancel(suite.ctx)
	cancel()

	data := map[string]interface{}{
		"email": "test@example.com",
		"age":   25,
		"name":  "John Doe",
	}

	result, err := suite.validationNode.Execute(cancelCtx, suite.wm, data)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "context cancelled")
}
