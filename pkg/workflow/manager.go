package workflow

import (
	"errors"
	"fmt"
	"sync"

	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/storage"
)

type WorkflowManager[T any] struct {
	graph      *Graph[T]
	stateStore storage.StateStore[T]
	mu         sync.Mutex
	tasks      map[string]func(T) (T, error)
	hooks      map[string]func(T) (T, error)
}

func NewWorkflowManager[T any]() *WorkflowManager[T] {
	return &WorkflowManager[T]{
		graph: &Graph[T]{},
		tasks: make(map[string]func(T) (T, error)),
		hooks: make(map[string]func(T) (T, error)),
	}
}

func (wm *WorkflowManager[T]) RegisterStateStore(store storage.StateStore[T]) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.stateStore = store
}

func (wm *WorkflowManager[T]) RegisterTask(name string, taskFunc func(T) (T, error)) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.tasks[name] = taskFunc
}

func (wm *WorkflowManager[T]) UnregisterTask(name string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.tasks, name)
}

func (wm *WorkflowManager[T]) RegisterHook(name string, hookFunc func(T) (T, error)) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.hooks[name] = hookFunc
}

func (wm *WorkflowManager[T]) UnregisterHook(name string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.hooks, name)
}

func (wm *WorkflowManager[T]) AddNode(node NodeInterface[T]) {
	wm.graph.Nodes = append(wm.graph.Nodes, node)
}

func (wm *WorkflowManager[T]) AddEdge(edge *Edge[T]) {
	wm.graph.Edges = append(wm.graph.Edges, edge)
}

func (wm *WorkflowManager[T]) Execute(startNodeID string, initialData T) (T, error) {
	startNode := wm.findNodeByID(startNodeID)
	if startNode == nil {
		return initialData, fmt.Errorf("start node not found with ID '%s'", startNodeID)
	}

	return wm.ExecuteNode(startNode, initialData)
}

func (wm *WorkflowManager[T]) LoadFromConfig(config *config.WorkflowConfig) error {
	nodeMap := make(map[string]NodeInterface[T])

	// First, create all nodes
	for _, nodeConfig := range config.Nodes {
		var node NodeInterface[T]
		switch nodeConfig.Type {
		case "Task":
			taskFunc, ok := wm.tasks[nodeConfig.TaskFunc]
			if !ok {
				return fmt.Errorf("task function not registered: %s", nodeConfig.TaskFunc)
			}

			var beforeFunc, afterFunc func(T) (T, error)
			if nodeConfig.BeforeExecute != "" {
				beforeFunc, ok = wm.hooks[nodeConfig.BeforeExecute]
				if !ok {
					return fmt.Errorf("before hook not registered: %s", nodeConfig.BeforeExecute)
				}
			}
			if nodeConfig.AfterExecute != "" {
				afterFunc, ok = wm.hooks[nodeConfig.AfterExecute]
				if !ok {
					return fmt.Errorf("after hook not registered: %s", nodeConfig.AfterExecute)
				}
			}

			node = &Node[T]{
				ID:            nodeConfig.ID,
				Type:          Task,
				TaskFunc:      taskFunc,
				BeforeExecute: beforeFunc,
				AfterExecute:  afterFunc,
			}

		case "Parallel":
			node = &ParallelNode[T]{
				Node: Node[T]{
					ID:   nodeConfig.ID,
					Type: Parallel,
				},
				// ParallelTasks will be assigned after all nodes are created
			}

		case "Branch":
			node = &BranchNode[T]{
				Node: Node[T]{
					ID:   nodeConfig.ID,
					Type: Branch,
				},
				// Branches will be assigned after all nodes are created
			}

		case "Foreach":
			taskFunc, ok := wm.tasks[nodeConfig.TaskFunc]
			if !ok {
				return fmt.Errorf("iterate function not registered: %s", nodeConfig.TaskFunc)
			}
			executionMode := Asynchronous
			if nodeConfig.ExecutionMode != "" {
				if nodeConfig.ExecutionMode == string(Synchronous) {
					executionMode = Synchronous
				} else if nodeConfig.ExecutionMode == string(Asynchronous) {
					executionMode = Asynchronous
				} else {
					return fmt.Errorf("invalid execution mode for ForeachNode: %s", nodeConfig.ExecutionMode)
				}
			}
			node = &ForeachNode[T]{
				Node: Node[T]{
					ID:   nodeConfig.ID,
					Type: Foreach,
				},
				IterateFunc:   taskFunc,
				ExecutionMode: executionMode,
			}

		case "Conditional":
			taskFunc, ok := wm.tasks[nodeConfig.TaskFunc]
			if !ok {
				return fmt.Errorf("condition function not registered: %s", nodeConfig.TaskFunc)
			}
			node = &ConditionalNode[T]{
				Node: Node[T]{
					ID:   nodeConfig.ID,
					Type: Conditional,
				},
				ConditionFunc: taskFunc,
			}

		default:
			return fmt.Errorf("unsupported node type: %s", nodeConfig.Type)
		}

		nodeMap[nodeConfig.ID] = node
	}

	// Now, set up connections and references
	for _, nodeConfig := range config.Nodes {
		node := nodeMap[nodeConfig.ID]
		switch n := node.(type) {
		case *ParallelNode[T]:
			parallelTasks := make([]NodeInterface[T], len(nodeConfig.ParallelTasks))
			for i, taskID := range nodeConfig.ParallelTasks {
				taskNode, ok := nodeMap[taskID]
				if !ok {
					return fmt.Errorf("parallel task node not found: %s", taskID)
				}
				parallelTasks[i] = taskNode
			}
			n.ParallelTasks = parallelTasks

		case *BranchNode[T]:
			branches := make([]NodeInterface[T], len(nodeConfig.Branches))
			for i, branchID := range nodeConfig.Branches {
				branchNode, ok := nodeMap[branchID]
				if !ok {
					return fmt.Errorf("branch node not found: %s", branchID)
				}
				branches[i] = branchNode
			}
			n.Branches = branches

		case *ConditionalNode[T]:
			trueNext, ok := nodeMap[nodeConfig.TrueNext]
			if !ok {
				return fmt.Errorf("trueNext node not found: %s", nodeConfig.TrueNext)
			}
			falseNext, ok := nodeMap[nodeConfig.FalseNext]
			if !ok {
				return fmt.Errorf("falseNext node not found: %s", nodeConfig.FalseNext)
			}
			n.TrueNext = trueNext
			n.FalseNext = falseNext
		}
	}

	// Add nodes to the graph
	for _, node := range nodeMap {
		wm.AddNode(node)
	}

	// Add edges
	for _, edgeConfig := range config.Edges {
		fromNode, ok := nodeMap[edgeConfig.From]
		if !ok {
			return fmt.Errorf("from node not found: %s", edgeConfig.From)
		}
		toNode, ok := nodeMap[edgeConfig.To]
		if !ok {
			return fmt.Errorf("to node not found: %s", edgeConfig.To)
		}
		wm.AddEdge(&Edge[T]{
			From: fromNode,
			To:   toNode,
		})
	}

	return nil
}

func (wm *WorkflowManager[T]) ExecuteNode(node NodeInterface[T], data T) (T, error) {
	if node == nil {
		return data, errors.New("ExecuteNode: node is nil")
	}

	// Load state if exists
	if wm.stateStore != nil {
		state, exists, err := wm.stateStore.LoadState(node.GetID())
		if err != nil {
			return data, err
		}
		if exists {
			data = state
		}
	}

	result, err := node.Execute(wm, data)
	if err != nil {
		return result, err
	}

	// Save state
	if wm.stateStore != nil {
		err = wm.stateStore.SaveState(node.GetID(), result)
		if err != nil {
			return result, err
		}
	}

	// Execute next nodes
	for _, edge := range wm.graph.Edges {
		if edge.From.GetID() == node.GetID() {
			if edge.To == nil {
				continue
			}
			if edge.Condition == nil || edge.Condition() {
				return wm.ExecuteNode(edge.To, result)
			}
		}
	}

	return result, nil
}

func (wm *WorkflowManager[T]) findNodeByID(id string) NodeInterface[T] {
	for _, node := range wm.graph.Nodes {
		if node.GetID() == id {
			return node
		}
	}
	return nil
}
