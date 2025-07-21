package unit

import (
	"context"
	"testing"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AdvancedValidationTestSuite struct {
	suite.Suite
	validator *workflow.AdvancedValidator
	ctx       context.Context
}

func (suite *AdvancedValidationTestSuite) SetupTest() {
	suite.validator = &workflow.AdvancedValidator{}
	suite.ctx = context.Background()
}

func TestAdvancedValidationTestSuite(t *testing.T) {
	suite.Run(t, new(AdvancedValidationTestSuite))
}

// Tests para ValidateTypes
func (suite *AdvancedValidationTestSuite) TestValidateTypesSuccess() {
	data := map[string]interface{}{
		"name":   "John Doe",
		"age":    30,
		"email":  "john@example.com",
		"active": true,
	}

	schema := map[string]workflow.TypeConstraint{
		"name": {
			Type:      "string",
			Required:  true,
			MinLength: 2,
			MaxLength: 50,
		},
		"age": {
			Type:     "int",
			Required: true,
			Min:      18,
			Max:      100,
		},
		"email": {
			Type:     "string",
			Required: true,
			Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		},
		"active": {
			Type:     "bool",
			Required: false,
		},
	}

	result, err := suite.validator.ValidateTypes(suite.ctx, data, schema)

	require.NoError(suite.T(), err)
	assert.True(suite.T(), result.Valid)
	assert.Empty(suite.T(), result.Errors)
	assert.Equal(suite.T(), 4, result.Summary["total_fields"])
	assert.Equal(suite.T(), 0, result.Summary["errors_count"])
}

func (suite *AdvancedValidationTestSuite) TestValidateTypesRequiredFieldMissing() {
	data := map[string]interface{}{
		"age": 30,
	}

	schema := map[string]workflow.TypeConstraint{
		"name": {
			Type:     "string",
			Required: true,
		},
		"age": {
			Type:     "int",
			Required: true,
		},
	}

	result, err := suite.validator.ValidateTypes(suite.ctx, data, schema)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.Len(suite.T(), result.Errors, 1)
	assert.Equal(suite.T(), "name", result.Errors[0].Field)
	assert.Equal(suite.T(), "required", result.Errors[0].Rule)
}

func (suite *AdvancedValidationTestSuite) TestValidateTypesInvalidType() {
	data := map[string]interface{}{
		"age": "not a number",
	}

	schema := map[string]workflow.TypeConstraint{
		"age": {
			Type:     "int",
			Required: true,
		},
	}

	result, err := suite.validator.ValidateTypes(suite.ctx, data, schema)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.Len(suite.T(), result.Errors, 1)
	assert.Equal(suite.T(), "age", result.Errors[0].Field)
}

func (suite *AdvancedValidationTestSuite) TestValidateTypesStringConstraints() {
	data := map[string]interface{}{
		"short":           "a",
		"long":            "this is a very long string that exceeds the maximum length allowed",
		"invalid_pattern": "invalid-email",
	}

	schema := map[string]workflow.TypeConstraint{
		"short": {
			Type:      "string",
			Required:  true,
			MinLength: 5,
		},
		"long": {
			Type:      "string",
			Required:  true,
			MaxLength: 10,
		},
		"invalid_pattern": {
			Type:     "string",
			Required: true,
			Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		},
	}

	result, err := suite.validator.ValidateTypes(suite.ctx, data, schema)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.Len(suite.T(), result.Errors, 3)
}

func (suite *AdvancedValidationTestSuite) TestValidateTypesContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data := map[string]interface{}{"name": "test"}
	schema := map[string]workflow.TypeConstraint{
		"name": {Type: "string", Required: true},
	}

	result, err := suite.validator.ValidateTypes(ctx, data, schema)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), context.Canceled, err)
	assert.NotNil(suite.T(), result)
}

// Tests para ValidateDependencies
func (suite *AdvancedValidationTestSuite) TestValidateDependenciesSuccess() {
	nodes := []workflow.NodeConfig{
		{ID: "node1", Dependencies: []string{}},
		{ID: "node2", Dependencies: []string{"node1"}},
		{ID: "node3", Dependencies: []string{"node2"}},
	}

	result, err := suite.validator.ValidateDependencies(suite.ctx, nodes)

	require.NoError(suite.T(), err)
	assert.True(suite.T(), result.Valid)
	assert.Empty(suite.T(), result.Errors)
}

func (suite *AdvancedValidationTestSuite) TestValidateDependenciesNotFound() {
	nodes := []workflow.NodeConfig{
		{ID: "node1", Dependencies: []string{"nonexistent"}},
		{ID: "node2", Dependencies: []string{"node1"}},
	}

	result, err := suite.validator.ValidateDependencies(suite.ctx, nodes)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.Len(suite.T(), result.Errors, 1)
	assert.Equal(suite.T(), "node1", result.Errors[0].Field)
	assert.Equal(suite.T(), "dependency_not_found", result.Errors[0].Rule)
}

func (suite *AdvancedValidationTestSuite) TestValidateDependenciesCircularDependency() {
	nodes := []workflow.NodeConfig{
		{ID: "node1", Dependencies: []string{"node3"}},
		{ID: "node2", Dependencies: []string{"node1"}},
		{ID: "node3", Dependencies: []string{"node2"}},
	}

	result, err := suite.validator.ValidateDependencies(suite.ctx, nodes)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)

	// Verificar que se detectó dependencia circular
	hasCircularError := false
	for _, err := range result.Errors {
		if err.Rule == "circular_dependency" {
			hasCircularError = true
			break
		}
	}
	assert.True(suite.T(), hasCircularError)
}

// Tests para DetectCycles
func (suite *AdvancedValidationTestSuite) TestDetectCyclesNoCycles() {
	graph := workflow.GraphDefinition{
		Nodes: []workflow.NodeConfig{
			{ID: "A", Dependencies: []string{}},
			{ID: "B", Dependencies: []string{"A"}},
			{ID: "C", Dependencies: []string{"B"}},
		},
		Edges: []workflow.EdgeDefinition{
			{From: "A", To: "B"},
			{From: "B", To: "C"},
		},
	}

	result, err := suite.validator.DetectCycles(suite.ctx, graph)

	require.NoError(suite.T(), err)
	assert.True(suite.T(), result.Valid)
	assert.Empty(suite.T(), result.Errors)
}

func (suite *AdvancedValidationTestSuite) TestDetectCyclesWithCycle() {
	graph := workflow.GraphDefinition{
		Nodes: []workflow.NodeConfig{
			{ID: "A", Dependencies: []string{"C"}},
			{ID: "B", Dependencies: []string{"A"}},
			{ID: "C", Dependencies: []string{"B"}},
		},
		Edges: []workflow.EdgeDefinition{
			{From: "A", To: "B"},
			{From: "B", To: "C"},
			{From: "C", To: "A"},
		},
	}

	result, err := suite.validator.DetectCycles(suite.ctx, graph)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)
}

// Tests para ValidateFunctionSignature
func (suite *AdvancedValidationTestSuite) TestValidateFunctionSignatureValid() {
	validFunc := func(ctx context.Context, data interface{}) (interface{}, error) {
		return data, nil
	}

	result, err := suite.validator.ValidateFunctionSignature(suite.ctx, validFunc)

	require.NoError(suite.T(), err)
	assert.True(suite.T(), result.Valid)
	assert.Empty(suite.T(), result.Errors)
}

func (suite *AdvancedValidationTestSuite) TestValidateFunctionSignatureInvalid() {
	invalidFunc := func(data string) string {
		return data
	}

	result, err := suite.validator.ValidateFunctionSignature(suite.ctx, invalidFunc)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)
}

func (suite *AdvancedValidationTestSuite) TestValidateFunctionSignatureNotFunction() {
	notAFunction := "this is not a function"

	result, err := suite.validator.ValidateFunctionSignature(suite.ctx, notAFunction)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)
	assert.Equal(suite.T(), "not_a_function", result.Errors[0].Rule)
}

// Tests para ValidateComplexSchema
func (suite *AdvancedValidationTestSuite) TestValidateComplexSchemaSuccess() {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"profile": map[string]interface{}{
				"age":    30,
				"active": true,
			},
		},
		"tags": []interface{}{"tag1", "tag2", "tag3"},
	}

	schema := workflow.ComplexSchema{
		Properties: map[string]workflow.PropertySchema{
			"user": {
				Type:     "object",
				Required: true,
				Properties: map[string]workflow.PropertySchema{
					"name": {
						Type:      "string",
						Required:  true,
						MinLength: 2,
					},
					"email": {
						Type:     "string",
						Required: true,
						Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
					},
					"profile": {
						Type:     "object",
						Required: true,
						Properties: map[string]workflow.PropertySchema{
							"age": {
								Type:     "int",
								Required: true,
								Min:      18,
							},
							"active": {
								Type: "bool",
							},
						},
					},
				},
			},
			"tags": {
				Type:     "array",
				Required: false,
				MinItems: 1,
				MaxItems: 5,
				Items:    &workflow.PropertySchema{Type: "string"},
			},
		},
	}

	result, err := suite.validator.ValidateComplexSchema(suite.ctx, data, schema)

	require.NoError(suite.T(), err)
	assert.True(suite.T(), result.Valid)
	assert.Empty(suite.T(), result.Errors)
}

func (suite *AdvancedValidationTestSuite) TestValidateComplexSchemaNestedErrors() {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "A",             // Too short
			"email": "invalid-email", // Invalid pattern
			"profile": map[string]interface{}{
				"age": 15, // Below minimum
			},
		},
		"tags": []interface{}{}, // Empty array but required
	}

	schema := workflow.ComplexSchema{
		Properties: map[string]workflow.PropertySchema{
			"user": {
				Type:     "object",
				Required: true,
				Properties: map[string]workflow.PropertySchema{
					"name": {
						Type:      "string",
						Required:  true,
						MinLength: 2,
					},
					"email": {
						Type:     "string",
						Required: true,
						Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
					},
					"profile": {
						Type:     "object",
						Required: true,
						Properties: map[string]workflow.PropertySchema{
							"age": {
								Type:     "int",
								Required: true,
								Min:      18,
							},
						},
					},
				},
			},
			"tags": {
				Type:     "array",
				Required: true,
				MinItems: 1,
			},
		},
	}

	result, err := suite.validator.ValidateComplexSchema(suite.ctx, data, schema)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)
	assert.True(suite.T(), len(result.Errors) >= 4) // At least 4 errors expected
}

// Tests para ValidatePerformanceConstraints
func (suite *AdvancedValidationTestSuite) TestValidatePerformanceConstraintsSuccess() {
	data := map[string]interface{}{
		"execution_time": 100,  // milliseconds
		"memory_usage":   50,   // MB
		"cpu_usage":      30.5, // percentage
		"timeout":        "5s", // duration string
	}

	result, err := suite.validator.ValidatePerformanceConstraints(suite.ctx, data)

	require.NoError(suite.T(), err)
	assert.True(suite.T(), result.Valid)
	assert.Empty(suite.T(), result.Errors)
}

func (suite *AdvancedValidationTestSuite) TestValidatePerformanceConstraintsViolation() {
	data := map[string]interface{}{
		"execution_time": 10000,     // Too high
		"memory_usage":   2048,      // Too high
		"cpu_usage":      150.0,     // Invalid percentage
		"timeout":        "invalid", // Invalid duration
	}

	result, err := suite.validator.ValidatePerformanceConstraints(suite.ctx, data)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)
}

// Test de conversión de datos
func (suite *AdvancedValidationTestSuite) TestValidateTypesDataConversion() {
	// Test with struct that needs JSON conversion
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := TestStruct{Name: "John", Age: 30}
	schema := map[string]workflow.TypeConstraint{
		"name": {Type: "string", Required: true},
		"age":  {Type: "int", Required: true},
	}

	result, err := suite.validator.ValidateTypes(suite.ctx, data, schema)

	require.NoError(suite.T(), err)
	assert.True(suite.T(), result.Valid)
	assert.Empty(suite.T(), result.Errors)
}

func (suite *AdvancedValidationTestSuite) TestValidateTypesInvalidData() {
	// Test with data that can't be converted
	data := make(chan int) // Channels can't be JSON marshaled
	schema := map[string]workflow.TypeConstraint{
		"test": {Type: "string", Required: true},
	}

	result, err := suite.validator.ValidateTypes(suite.ctx, data, schema)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)
	assert.Equal(suite.T(), "type_conversion", result.Errors[0].Rule)
}
