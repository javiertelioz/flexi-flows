package unit

import (
	"context"
	"testing"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type TransformNodeTestSuite struct {
	suite.Suite
	wm            *workflow.WorkflowManager
	transformNode *workflow.TransformNode
	ctx           context.Context
}

func TestTransformNodeTestSuite(t *testing.T) {
	suite.Run(t, new(TransformNodeTestSuite))
}

func (suite *TransformNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()
	suite.ctx = context.Background()

	transformType, err := workflow.ParseNodeType("transform")
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.transformNode = &workflow.TransformNode{
		Node: workflow.Node[interface{}]{
			ID:   "transform_test",
			Type: transformType,
		},
		KeepOriginal: false,
	}
}

func (suite *TransformNodeTestSuite) TestTransformNodeExecuteSimpleMapping() {
	suite.transformNode.Mapping = map[string]interface{}{
		"full_name": "name",
		"user_age":  "age",
		"constant":  "fixed_value",
	}

	data := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
		"city": "New York",
	}

	result, err := suite.transformNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map")

	suite.Equal("John Doe", resultMap["full_name"])
	suite.Equal(30, resultMap["user_age"])
	suite.Equal("fixed_value", resultMap["constant"])
	suite.NotContains(resultMap, "city") // Should not have original field
}

func (suite *TransformNodeTestSuite) TestTransformNodeExecuteWithKeepOriginal() {
	suite.transformNode.KeepOriginal = true
	suite.transformNode.Mapping = map[string]interface{}{
		"full_name": "name",
	}

	data := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	result, err := suite.transformNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map")

	suite.Equal("John Doe", resultMap["full_name"])
	suite.Equal("John Doe", resultMap["name"]) // Original kept
	suite.Equal(30, resultMap["age"])          // Original kept
}

func (suite *TransformNodeTestSuite) TestTransformNodeExecuteTransformRules() {
	rules := []workflow.TransformRule{
		{
			SourceField: "name",
			TargetField: "upper_name",
			Operation:   "uppercase",
		},
		{
			SourceField: "email",
			TargetField: "lower_email",
			Operation:   "lowercase",
		},
		{
			SourceField: "age",
			TargetField: "age_plus_ten",
			Operation:   "calculate",
			Parameters: map[string]interface{}{
				"operation": "add",
				"operand":   10.0,
			},
		},
	}

	suite.transformNode.Rules = rules

	data := map[string]interface{}{
		"name":  "john doe",
		"email": "JOHN@EXAMPLE.COM",
		"age":   25,
	}

	result, err := suite.transformNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map")

	suite.Equal("JOHN DOE", resultMap["upper_name"])
	suite.Equal("john@example.com", resultMap["lower_email"])
	suite.Equal(35.0, resultMap["age_plus_ten"])
}

func (suite *TransformNodeTestSuite) TestTransformNodeExecuteCustomFunction() {
	// Función personalizada que duplica todos los valores numéricos
	customFunc := func(data interface{}) (interface{}, error) {
		dataMap := data.(map[string]interface{})
		result := make(map[string]interface{})

		for key, value := range dataMap {
			if num, ok := value.(int); ok {
				result[key] = num * 2
			} else {
				result[key] = value
			}
		}

		return result, nil
	}

	suite.transformNode.CustomFunc = customFunc

	data := map[string]interface{}{
		"count":  5,
		"amount": 100,
		"name":   "test",
	}

	result, err := suite.transformNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map")

	suite.Equal(10, resultMap["count"])    // 5 * 2
	suite.Equal(200, resultMap["amount"])  // 100 * 2
	suite.Equal("test", resultMap["name"]) // unchanged
}

func (suite *TransformNodeTestSuite) TestTransformNodeExecuteStringOperations() {
	rules := []workflow.TransformRule{
		{
			SourceField: "text",
			TargetField: "trimmed",
			Operation:   "trim",
		},
		{
			SourceField: "text",
			TargetField: "replaced",
			Operation:   "replace",
			Parameters: map[string]interface{}{
				"old": "world",
				"new": "universe",
			},
		},
		{
			SourceField: "text",
			TargetField: "split_words",
			Operation:   "split",
			Parameters: map[string]interface{}{
				"separator": " ",
			},
		},
	}

	suite.transformNode.Rules = rules

	data := map[string]interface{}{
		"text": "  hello world  ",
	}

	result, err := suite.transformNode.Execute(suite.ctx, suite.wm, data)

	suite.NoError(err)
	suite.NotNil(result)

	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map")

	suite.Equal("hello world", resultMap["trimmed"])
	suite.Equal("  hello universe  ", resultMap["replaced"])

	splitResult, ok := resultMap["split_words"].([]string)
	suite.True(ok, "Split result should be []string")
	suite.Equal([]string{"", "", "hello", "world", "", ""}, splitResult)
}

func (suite *TransformNodeTestSuite) TestTransformNodeExecuteContextCancellation() {
	cancelCtx, cancel := context.WithCancel(suite.ctx)
	cancel()

	data := map[string]interface{}{
		"name": "test",
	}

	result, err := suite.transformNode.Execute(cancelCtx, suite.wm, data)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "context cancelled")
}
