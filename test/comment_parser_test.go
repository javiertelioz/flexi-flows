package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type CommentParserTestSuite struct {
	suite.Suite
}

func TestCommentParserTestSuite(t *testing.T) {
	suite.Run(t, new(CommentParserTestSuite))
}

func (suite *CommentParserTestSuite) SetupTest() {
	// No setup needed for these tests
}

// Given methods
func (suite *CommentParserTestSuite) givenTestSourceCode() string {
	return `
		// isPrimeFunc
		// @type: task
		// @description checks if a given integer is a prime number.
		// @input data (int): The integer to be checked for primality.
		// @output bool: Returns true if the input integer is a prime number, false otherwise.
		// @output error: Returns an error if the input integer is less than 1.
		func isPrimeFunc(data int) (bool, error) {
			if data <= 1 {
				return false, nil
			}
			for i := 2; i <= int(math.Sqrt(float64(data))); i++ {
				if data%i == 0 {
					return false, nil
				}
			}
			return true, nil
		}

		// beforeIsPrime
		// @type: pre-hook
		// @before isPrimeFunc
		// @description execute before isPrimeFunc.
		// @input data (int): The integer to be checked for primality.
		// @output int: Returns integer is a prime number.
		// @output error: Returns an error if occurs.
		func beforeIsPrime(data int) (int, error) {
			fmt.Println("Before isPrime: received data", data)
			return data, nil
		}

		// afterIsPrime
		// @type: post-hook
		// @after isPrimeFunc
		// @description execute after isPrimeFunc.
		// @input result (bool): The result of the isPrimeFunc.
		// @output bool: Returns the result after post processing.
		// @output error: Returns an error if occurs.
		func afterIsPrime(result bool) (bool, error) {
			fmt.Println("After isPrime: result is", result)
			return result, nil
		}
	`
}

// When methods
func (suite *CommentParserTestSuite) whenParseComments(src string) (map[string]workflow.FunctionMetadata, error) {
	dir := "./test_src"
	err := os.Mkdir(dir, 0755)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	err = os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0644)
	if err != nil {
		return nil, err
	}

	return workflow.ParseComments(dir)
}

// Then methods
func (suite *CommentParserTestSuite) thenExpectMetadataToMatch(expected, actual map[string]workflow.FunctionMetadata) {
	suite.Equal(expected, actual)
}

func (suite *CommentParserTestSuite) thenExpectNoError(err error) {
	suite.NoError(err)
}

// Test methods
/*func (suite *CommentParserTestSuite) TestParseCommentsWithValidSource() {
	// Given
	src := suite.givenTestSourceCode()
	expected := map[string]workflow.FunctionMetadata{
		"isPrimeFunc": {
			Type:        "task",
			Description: "checks if a given integer is a prime number.",
			Parameters: []workflow.ParameterMetadata{
				{Type: "int", Description: "The integer to be checked for primality."},
			},
			Returns: []workflow.ReturnMetadata{
				{Type: "bool", Description: "Returns true if the input integer is a prime number, false otherwise."},
				{Type: "error", Description: "Returns an error if the input integer is less than 1."},
			},
			Hooks: map[string]workflow.HookMetadata{
				"pre-hook_beforeIsPrime": {
					Type:        "pre-hook",
					Description: "execute before isPrimeFunc.",
					Parameters: []workflow.ParameterMetadata{
						{Type: "int", Description: "The integer to be checked for primality."},
					},
					Returns: []workflow.ReturnMetadata{
						{Type: "int", Description: "Returns integer is a prime number."},
						{Type: "error", Description: "Returns an error if occurs."},
					},
				},
				"post-hook_afterIsPrime": {
					Type:        "post-hook",
					Description: "execute after isPrimeFunc.",
					Parameters: []workflow.ParameterMetadata{
						{Type: "bool", Description: "The result of the isPrimeFunc."},
					},
					Returns: []workflow.ReturnMetadata{
						{Type: "bool", Description: "Returns the result after post processing."},
						{Type: "error", Description: "Returns an error if occurs."},
					},
				},
			},
		},
	}

	// When
	actual, err := suite.whenParseComments(src)

	// Then
	suite.thenExpectNoError(err)
	suite.thenExpectMetadataToMatch(expected, actual)
}*/

func (suite *CommentParserTestSuite) TestParseCommentsWithInvalidSource() {
	// Given
	src := "invalid source code"

	// When
	_, err := suite.whenParseComments(src)

	// Then
	suite.Error(err)
}
