package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/stretchr/testify/suite"
)

type HTTPNodeTestSuite struct {
	suite.Suite
	wm       *workflow.WorkflowManager
	httpNode *workflow.HTTPNode
	ctx      context.Context
	server   *httptest.Server
}

func TestHTTPNodeTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPNodeTestSuite))
}

func (suite *HTTPNodeTestSuite) SetupTest() {
	suite.wm = workflow.NewWorkflowManager()
	suite.ctx = context.Background()

	// Crear servidor de prueba
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success", "data": "test"}`))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		case "/echo":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if r.Method == "POST" {
				w.Write([]byte(`{"received": "post data"}`))
			} else {
				w.Write([]byte(`{"method": "` + r.Method + `"}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found"}`))
		}
	}))

	httpType, err := workflow.ParseNodeType("http")
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.httpNode = &workflow.HTTPNode{
		Node: workflow.Node[interface{}]{
			ID:   "http_test",
			Type: httpType,
		},
		URL:     suite.server.URL + "/success",
		Method:  "GET",
		Timeout: 5 * time.Second,
	}
}

func (suite *HTTPNodeTestSuite) TearDownTest() {
	if suite.server != nil {
		suite.server.Close()
	}
}

func (suite *HTTPNodeTestSuite) TestHTTPNodeExecuteGETSuccess() {
	result, err := suite.httpNode.Execute(suite.ctx, suite.wm, nil)

	suite.NoError(err)
	suite.NotNil(result)

	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map")

	suite.Equal(200, resultMap["status_code"])
	suite.True(resultMap["success"].(bool))
	suite.Contains(resultMap["body"].(string), "success")
}

func (suite *HTTPNodeTestSuite) TestHTTPNodeExecutePOSTWithBody() {
	suite.httpNode.URL = suite.server.URL + "/echo"
	suite.httpNode.Method = "POST"
	suite.httpNode.Body = map[string]interface{}{
		"test":   "data",
		"number": 123,
	}
	suite.httpNode.Headers = map[string]string{
		"Custom-Header": "test-value",
	}

	result, err := suite.httpNode.Execute(suite.ctx, suite.wm, nil)

	suite.NoError(err)
	suite.NotNil(result)

	resultMap, ok := result.(map[string]interface{})
	suite.True(ok, "Result should be a map")

	suite.Equal(200, resultMap["status_code"])
	suite.True(resultMap["success"].(bool))
}

func (suite *HTTPNodeTestSuite) TestHTTPNodeExecuteErrorResponse() {
	suite.httpNode.URL = suite.server.URL + "/error"

	result, err := suite.httpNode.Execute(suite.ctx, suite.wm, nil)

	// El HTTPNode retorna error cuando el status code >= 400
	suite.Error(err)
	// El resultado es nil cuando hay error HTTP
	suite.Nil(result)

	// Verificar que el error contiene información sobre el status code
	suite.Contains(err.Error(), "500")
	suite.Contains(err.Error(), "HTTP request failed")
}

func (suite *HTTPNodeTestSuite) TestHTTPNodeExecuteInvalidURL() {
	suite.httpNode.URL = "invalid-url"

	result, err := suite.httpNode.Execute(suite.ctx, suite.wm, nil)

	suite.Error(err)
	suite.Nil(result)
}

func (suite *HTTPNodeTestSuite) TestHTTPNodeExecuteContextCancellation() {
	// Crear contexto que se cancela inmediatamente
	cancelCtx, cancel := context.WithCancel(suite.ctx)
	cancel()

	result, err := suite.httpNode.Execute(cancelCtx, suite.wm, nil)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "context cancelled")
}
