package workflow

import "context"

type ConditionalNode struct {
	Node[interface{}]
	Condition func(data interface{}) bool
	TrueNext  NodeInterface
	FalseNext NodeInterface
}

func (n *ConditionalNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(n.ID, n.Type, "context cancelled before conditional execution", ctx.Err())
	default:
	}

	// Evaluar la condición
	var conditionResult bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Si la condición causa panic, tratarlo como false
				conditionResult = false
			}
		}()
		conditionResult = n.Condition(data)
	}()

	// Ejecutar la rama correspondiente
	if conditionResult {
		if n.TrueNext != nil {
			return wm.ExecuteNodeWithContext(ctx, n.TrueNext, data)
		}
		return map[string]interface{}{
			"condition_result": true,
			"data":             data,
			"message":          "condition was true but no true_next node defined",
		}, nil
	} else {
		if n.FalseNext != nil {
			return wm.ExecuteNodeWithContext(ctx, n.FalseNext, data)
		}
		return map[string]interface{}{
			"condition_result": false,
			"data":             data,
			"message":          "condition was false but no false_next node defined",
		}, nil
	}
}
