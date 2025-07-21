package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPNode representa un nodo que realiza peticiones HTTP
type HTTPNode struct {
	Node[interface{}]
	URL     string
	Method  string
	Headers map[string]string
	Body    interface{}
	Timeout time.Duration
	Client  *http.Client
}

// Execute realiza la petición HTTP y retorna la respuesta
func (hn *HTTPNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(hn.ID, hn.Type, "context cancelled before HTTP request", ctx.Err())
	default:
	}

	// Configurar cliente HTTP si no existe
	if hn.Client == nil {
		timeout := hn.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second // timeout por defecto
		}
		hn.Client = &http.Client{
			Timeout: timeout,
		}
	}

	// Preparar el body de la petición
	var requestBody io.Reader
	if hn.Body != nil {
		// Si el body es un string, usarlo directamente
		if bodyStr, ok := hn.Body.(string); ok {
			requestBody = bytes.NewBufferString(bodyStr)
		} else {
			// Si no, serializar como JSON
			bodyBytes, err := json.Marshal(hn.Body)
			if err != nil {
				return nil, NewWorkflowError(hn.ID, hn.Type, "failed to marshal request body", err)
			}
			requestBody = bytes.NewBuffer(bodyBytes)
		}
	}

	// Crear la petición HTTP
	req, err := http.NewRequestWithContext(ctx, hn.Method, hn.URL, requestBody)
	if err != nil {
		return nil, NewWorkflowError(hn.ID, hn.Type, "failed to create HTTP request", err)
	}

	// Agregar headers
	for key, value := range hn.Headers {
		req.Header.Set(key, value)
	}

	// Si no hay Content-Type y hay body, usar JSON por defecto
	if hn.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Realizar la petición
	resp, err := hn.Client.Do(req)
	if err != nil {
		return nil, NewWorkflowError(hn.ID, hn.Type, "HTTP request failed", err)
	}
	defer resp.Body.Close()

	// Leer la respuesta
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewWorkflowError(hn.ID, hn.Type, "failed to read response body", err)
	}

	// Preparar el resultado
	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(responseBody),
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	// Si la respuesta es JSON, intentar parsearla
	if resp.Header.Get("Content-Type") == "application/json" ||
		(len(responseBody) > 0 && responseBody[0] == '{') {
		var jsonBody interface{}
		if err := json.Unmarshal(responseBody, &jsonBody); err == nil {
			result["json"] = jsonBody
		}
	}

	// Si el status code indica error, retornar error
	if resp.StatusCode >= 400 {
		return result, NewWorkflowError(hn.ID, hn.Type,
			fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode),
			fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(responseBody)))
	}

	return result, nil
}
