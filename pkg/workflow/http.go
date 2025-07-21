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

	// Validar URL
	if hn.URL == "" {
		return nil, NewWorkflowError(hn.ID, hn.Type, "URL is required for HTTP node", fmt.Errorf("empty URL"))
	}

	// Crear cliente HTTP
	client := &http.Client{
		Timeout: hn.Timeout,
	}

	// Crear request
	req, err := http.NewRequestWithContext(ctx, hn.Method, hn.URL, nil)
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

	// Preparar el body de la petición si existe
	if hn.Body != nil {
		// Si el body es un string, usarlo directamente
		if bodyStr, ok := hn.Body.(string); ok {
			req.Body = io.NopCloser(bytes.NewBufferString(bodyStr))
		} else {
			// Si no, serializar como JSON
			bodyBytes, err := json.Marshal(hn.Body)
			if err != nil {
				return nil, NewWorkflowError(hn.ID, hn.Type, "failed to marshal request body", err)
			}
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// Ejecutar request
	resp, err := client.Do(req)
	if err != nil {
		return nil, NewWorkflowError(hn.ID, hn.Type, "HTTP request failed", err)
	}
	defer resp.Body.Close()

	// Leer response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewWorkflowError(hn.ID, hn.Type, "failed to read response body", err)
	}

	// Preparar el resultado
	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(body),
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	// Si la respuesta es JSON, intentar parsearla
	if resp.Header.Get("Content-Type") == "application/json" ||
		(len(body) > 0 && body[0] == '{') {
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			result["json"] = jsonBody
		}
	}

	// Verificar status code
	if resp.StatusCode >= 400 {
		return nil, NewWorkflowError(hn.ID, hn.Type,
			fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode),
			fmt.Errorf("status: %s, body: %s", resp.Status, string(body)))
	}

	return result, nil
}
