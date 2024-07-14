package workflow

import (
	"reflect"
	"sync"

	"errors"

	"github.com/javiertelioz/go-flows/pkg/workflow/config"
	"github.com/javiertelioz/go-flows/pkg/workflow/storage"
)

type WorkflowManager struct {
	graph      *Graph
	stateStore storage.StateStore
	mu         sync.Mutex
	tasks      map[string]interface{}
	hooks      map[string]interface{}
}

func NewWorkflowManager(stateStore storage.StateStore, tasks, hooks map[string]interface{}) *WorkflowManager {
	return &WorkflowManager{
		graph:      &Graph{},
		stateStore: stateStore,
		tasks:      tasks,
		hooks:      hooks,
	}
}

func (wm *WorkflowManager) AddNode(node NodeInterface) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.graph.Nodes = append(wm.graph.Nodes, node)
}

func (wm *WorkflowManager) AddEdge(edge *Edge) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.graph.Edges = append(wm.graph.Edges, edge)
}

func (wm *WorkflowManager) Execute(startNodeID string, initialData interface{}) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	startNode := wm.findNodeByID(startNodeID)
	if startNode == nil {
		return errors.New("start node not found")
	}

	_, err := wm.ExecuteNode(startNode, initialData)
	return err
}

func (wm *WorkflowManager) LoadFromConfig(config *config.WorkflowConfig) error {
	nodeMap := make(map[string]NodeInterface)

	for _, nodeConfig := range config.Nodes {
		var node NodeInterface
		switch nodeConfig.Type {
		case "Task":
			taskFunc, ok := wm.tasks[nodeConfig.TaskFunc]
			if !ok {
				return errors.New("tasks function not found: " + nodeConfig.TaskFunc)
			}
			var beforeFunc, afterFunc func(interface{}) (interface{}, error)
			if nodeConfig.BeforeExecute != "" {
				beforeFunc = wrapHookFunc(wm.hooks[nodeConfig.BeforeExecute])
			}
			if nodeConfig.AfterExecute != "" {
				afterFunc = wrapHookFunc(wm.hooks[nodeConfig.AfterExecute])
			}
			node = &Node[interface{}]{
				ID:            nodeConfig.ID,
				Type:          Task,
				TaskFunc:      wrapTaskFunc(taskFunc),
				BeforeExecute: beforeFunc,
				AfterExecute:  afterFunc,
			}
		case "Parallel":
			parallelTasks := make([]NodeInterface, len(nodeConfig.ParallelTasks))
			for i, taskID := range nodeConfig.ParallelTasks {
				taskNode, ok := nodeMap[taskID]
				if !ok {
					return errors.New("tasks node not found: " + taskID)
				}
				parallelTasks[i] = taskNode
			}
			node = &ParallelNode{
				Node: Node[interface{}]{
					ID:   nodeConfig.ID,
					Type: Branch,
				},
				ParallelTasks: parallelTasks,
			}
		default:
			return errors.New("unsupported node type: " + nodeConfig.Type)
		}

		nodeMap[nodeConfig.ID] = node
		wm.AddNode(node)
	}

	for _, edgeConfig := range config.Edges {
		fromNode, ok := nodeMap[edgeConfig.From]
		if !ok {
			return errors.New("from node not found: " + edgeConfig.From)
		}
		toNode, ok := nodeMap[edgeConfig.To]
		if !ok {
			return errors.New("to node not found: " + edgeConfig.To)
		}
		wm.AddEdge(&Edge{
			From: fromNode,
			To:   toNode,
		})
	}

	return nil
}

func (wm *WorkflowManager) ExecuteNode(node NodeInterface, data interface{}) (interface{}, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}

	if wm.stateStore != nil {
		state, err := wm.stateStore.LoadState(node.GetID())
		if err != nil {
			return nil, err
		}
		if state != nil {
			data = state
		}
	}

	result, err := node.Execute(wm, data)
	if err != nil {
		return nil, err
	}

	if wm.stateStore != nil {
		err = wm.stateStore.SaveState(node.GetID(), result)
		if err != nil {
			return nil, err
		}
	}

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

func (wm *WorkflowManager) findNodeByID(id string) NodeInterface {
	for _, node := range wm.graph.Nodes {
		if node.GetID() == id {
			return node
		}
	}
	return nil
}

func wrapTaskFunc(taskFunc interface{}) func(interface{}) (interface{}, error) {
	return func(data interface{}) (interface{}, error) {
		result := reflect.ValueOf(taskFunc).Call([]reflect.Value{reflect.ValueOf(data)})
		if len(result) == 2 && !result[1].IsNil() {
			return result[0].Interface(), result[1].Interface().(error)
		}
		return result[0].Interface(), nil
	}
}

func wrapHookFunc(hookFunc interface{}) func(interface{}) (interface{}, error) {
	return func(data interface{}) (interface{}, error) {
		result := reflect.ValueOf(hookFunc).Call([]reflect.Value{reflect.ValueOf(data)})
		if len(result) == 2 && !result[1].IsNil() {
			return result[0].Interface(), result[1].Interface().(error)
		}
		return result[0].Interface(), nil
	}
}
