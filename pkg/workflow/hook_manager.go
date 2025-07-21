package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultHookManager implementa HookManager
type DefaultHookManager struct {
	hooks map[HookType][]Hook
	mu    sync.RWMutex
}

// NewHookManager crea un nuevo manager de hooks
func NewHookManager() *DefaultHookManager {
	return &DefaultHookManager{
		hooks: make(map[HookType][]Hook),
	}
}

// RegisterHook registra un hook para un tipo específico
func (hm *DefaultHookManager) RegisterHook(hookType HookType, hook Hook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hooks[hookType] = append(hm.hooks[hookType], hook)
}

// ExecuteHooks ejecuta todos los hooks de un tipo específico
func (hm *DefaultHookManager) ExecuteHooks(ctx context.Context, hookType HookType, hookCtx *HookContext) error {
	hm.mu.RLock()
	hooks := hm.hooks[hookType]
	hm.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.Execute(ctx, hookCtx); err != nil {
			return NewWorkflowError("hook_manager", Task, "hook execution failed", err)
		}
	}
	return nil
}

// FunctionHook adapta una función para ser un Hook
type FunctionHook struct {
	Func interface{}
}

// NewFunctionHook crea un hook desde una función
func NewFunctionHook(fn interface{}) (*FunctionHook, error) {
	// Validar que sea una función con la firma correcta
	if err := validateHookFunction(fn); err != nil {
		return nil, err
	}
	return &FunctionHook{Func: fn}, nil
}

// Execute ejecuta la función de hook
func (fh *FunctionHook) Execute(ctx context.Context, hookCtx *HookContext) error {
	return executeHookFunction(ctx, fh.Func, hookCtx)
}

// LoggingHook es un hook que registra información
type LoggingHook struct {
	Logger Logger
	Level  LogLevel
}

type Logger interface {
	Log(level LogLevel, message string, fields map[string]interface{})
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Execute ejecuta el hook de logging
func (lh *LoggingHook) Execute(ctx context.Context, hookCtx *HookContext) error {
	fields := map[string]interface{}{
		"node_id":      hookCtx.NodeID,
		"node_type":    hookCtx.NodeType,
		"hook_type":    hookCtx.HookType,
		"execution_id": hookCtx.ExecutionID,
		"timestamp":    hookCtx.Timestamp,
	}

	message := fmt.Sprintf("Hook executed: %s on %s", hookCtx.HookType, hookCtx.NodeID)
	lh.Logger.Log(lh.Level, message, fields)
	return nil
}

// MetricsHook recopila métricas de ejecución
type MetricsHook struct {
	Collector MetricsCollector
}

type MetricsCollector interface {
	RecordHookExecution(hookType HookType, nodeType NodeType, duration time.Duration)
	RecordNodeExecution(nodeID string, nodeType NodeType, duration time.Duration, success bool)
	IncrementCounter(name string, tags map[string]string)
}

// Execute ejecuta el hook de métricas
func (mh *MetricsHook) Execute(ctx context.Context, hookCtx *HookContext) error {
	if mh.Collector == nil {
		return nil
	}

	duration := time.Since(hookCtx.Timestamp)
	mh.Collector.RecordHookExecution(hookCtx.HookType, hookCtx.NodeType, duration)

	tags := map[string]string{
		"node_type": hookCtx.NodeType.String(),
		"hook_type": string(hookCtx.HookType),
	}
	mh.Collector.IncrementCounter("hooks_executed", tags)
	return nil
}

// ConditionalHook ejecuta un hook solo si se cumple una condición
type ConditionalHook struct {
	Condition func(ctx context.Context, hookCtx *HookContext) bool
	Hook      Hook
}

// Execute ejecuta el hook condicionalmente
func (ch *ConditionalHook) Execute(ctx context.Context, hookCtx *HookContext) error {
	if ch.Condition != nil && !ch.Condition(ctx, hookCtx) {
		return nil
	}
	return ch.Hook.Execute(ctx, hookCtx)
}

// ChainHook ejecuta múltiples hooks en secuencia
type ChainHook struct {
	Hooks []Hook
}

// Execute ejecuta todos los hooks en la cadena
func (ch *ChainHook) Execute(ctx context.Context, hookCtx *HookContext) error {
	for i, hook := range ch.Hooks {
		if err := hook.Execute(ctx, hookCtx); err != nil {
			return fmt.Errorf("chain hook failed at position %d: %w", i, err)
		}
	}
	return nil
}

// AsyncHook ejecuta un hook de forma asíncrona
type AsyncHook struct {
	Hook    Hook
	Timeout time.Duration
}

// Execute ejecuta el hook de forma asíncrona
func (ah *AsyncHook) Execute(ctx context.Context, hookCtx *HookContext) error {
	if ah.Timeout == 0 {
		ah.Timeout = 30 * time.Second
	}

	asyncCtx, cancel := context.WithTimeout(ctx, ah.Timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- ah.Hook.Execute(asyncCtx, hookCtx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-asyncCtx.Done():
		return NewWorkflowError(hookCtx.NodeID, hookCtx.NodeType,
			"async hook timeout", asyncCtx.Err())
	}
}
