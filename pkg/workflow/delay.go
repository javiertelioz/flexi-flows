package workflow

import (
	"context"
	"time"
)

// DelayNode representa un nodo que introduce un retraso temporal
type DelayNode struct {
	Node[interface{}]
	Duration time.Duration
}

// Execute introduce un retraso y luego retorna los datos sin modificar
func (dn *DelayNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado antes del delay
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(dn.ID, dn.Type, "context cancelled before delay", ctx.Err())
	default:
	}

	// Crear un timer con la duración especificada
	timer := time.NewTimer(dn.Duration)
	defer timer.Stop()

	// Esperar hasta que el timer expire o el contexto sea cancelado
	select {
	case <-timer.C:
		// El delay se completó exitosamente
		return map[string]interface{}{
			"message":  "delay completed successfully",
			"duration": dn.Duration.String(),
			"data":     data,
		}, nil

	case <-ctx.Done():
		// El contexto fue cancelado durante el delay
		return nil, NewWorkflowError(dn.ID, dn.Type, "context cancelled during delay", ctx.Err())
	}
}
