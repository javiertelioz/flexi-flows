package test

import (
	"os"
	"testing"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type StateStoreTestSuite struct {
	suite.Suite
	jsonFilePath string
	jsonStore    *workflow.JSONStateStore
	memStore     *workflow.MemoryStateStore
}

func TestStateStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StateStoreTestSuite))
}

func (suite *StateStoreTestSuite) SetupTest() {
	// Set up JSONStateStore
	suite.jsonFilePath = "test_state.json"
	suite.jsonStore = workflow.NewJSONStateStore(suite.jsonFilePath)

	// Set up MemoryStateStore
	suite.memStore = workflow.NewMemoryStateStore()
}

func (suite *StateStoreTestSuite) TearDownTest() {
	os.Remove(suite.jsonFilePath)
}

// Given methods for JSONStateStore
func (suite *StateStoreTestSuite) givenJSONStoreWithState(nodeID string, data interface{}) {
	suite.jsonStore.SaveState(nodeID, data)
}

// Given methods for MemoryStateStore
func (suite *StateStoreTestSuite) givenMemoryStoreWithState(nodeID string, data interface{}) {
	suite.memStore.SaveState(nodeID, data)
}

// When methods for JSONStateStore
func (suite *StateStoreTestSuite) whenLoadStateFromJSONStore(nodeID string) (interface{}, error) {
	return suite.jsonStore.LoadState(nodeID)
}

// When methods for MemoryStateStore
func (suite *StateStoreTestSuite) whenLoadStateFromMemoryStore(nodeID string) (interface{}, error) {
	return suite.memStore.LoadState(nodeID)
}

// Then methods for both stores
func (suite *StateStoreTestSuite) thenExpectStateToMatch(expected, actual interface{}, err error) {
	suite.NoError(err)
	suite.Equal(expected, actual)
}

func (suite *StateStoreTestSuite) thenExpectNoState(err error) {
	suite.NoError(err)
}

func (suite *StateStoreTestSuite) thenExpectError(err error) {
	suite.Error(err)
}

// JSONStateStore Tests
func (suite *StateStoreTestSuite) TestJSONStoreSaveAndLoadState() {
	nodeID := "testNode"
	data := "testData"

	// Given
	suite.givenJSONStoreWithState(nodeID, data)

	// When
	loadedData, err := suite.whenLoadStateFromJSONStore(nodeID)

	// Then
	suite.thenExpectStateToMatch(data, loadedData, err)
}

func (suite *StateStoreTestSuite) TestJSONStoreLoadStateNotFound() {
	nodeID := "nonExistentNode"

	// When
	loadedData, err := suite.whenLoadStateFromJSONStore(nodeID)

	// Then
	suite.thenExpectNoState(err)
	suite.Nil(loadedData)
}

func (suite *StateStoreTestSuite) TestJSONStoreLoadStateInvalidFile() {
	// Create an invalid JSON file
	os.WriteFile(suite.jsonFilePath, []byte("invalid json"), 0644)

	// When
	_, err := suite.whenLoadStateFromJSONStore("testNode")

	// Then
	suite.thenExpectError(err)
}

// MemoryStateStore Tests
func (suite *StateStoreTestSuite) TestMemoryStoreSaveAndLoadState() {
	nodeID := "testNode"
	data := "testData"

	// Given
	suite.givenMemoryStoreWithState(nodeID, data)

	// When
	loadedData, err := suite.whenLoadStateFromMemoryStore(nodeID)

	// Then
	suite.thenExpectStateToMatch(data, loadedData, err)
}

func (suite *StateStoreTestSuite) TestMemoryStoreLoadStateNotFound() {
	nodeID := "nonExistentNode"

	// When
	loadedData, err := suite.whenLoadStateFromMemoryStore(nodeID)

	// Then
	suite.thenExpectNoState(err)
	suite.Nil(loadedData)
}
