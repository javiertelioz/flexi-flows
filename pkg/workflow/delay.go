package workflow

import (
	"context"
	"fmt"
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
		return nil, NewWorkflowError(dn.ID, dn.Type, "context cancelled before delay execution", ctx.Err())
	default:
	}

	// Crear un timer con la duración especificada
	timer := time.NewTimer(dn.Duration)
	defer timer.Stop()

	startTime := time.Now()
	// Esperar hasta que el timer expire o el contexto sea cancelado
	select {
	case <-ctx.Done():
		// El contexto fue cancelado durante el delay
		return nil, NewWorkflowError(dn.ID, dn.Type,
			fmt.Sprintf("context cancelled during delay (waited %v of %v)", time.Since(startTime), dn.Duration),
			ctx.Err())
	case <-timer.C:
		// El delay se completó exitosamente
		return map[string]interface{}{
			"message":  "delay completed successfully",
			"duration": dn.Duration.String(),
			"data":     data,
		}, nil
	}
}
