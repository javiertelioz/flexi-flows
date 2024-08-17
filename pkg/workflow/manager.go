package workflow

import (
	"fmt"
	"reflect"
	"sync"

	"errors"

	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/storage"
)

type WorkflowManager struct {
	graph      *Graph
	stateStore storage.StateStore
	mu         sync.Mutex
	tasks      map[string]interface{}
	hooks      map[string]interface{}
}

func NewWorkflowManager() *WorkflowManager {
	return &WorkflowManager{
		graph: &Graph{},
		tasks: make(map[string]interface{}),
		hooks: make(map[string]interface{}),
	}
}

func (wm *WorkflowManager) RegisterStateStore(store storage.StateStore) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.stateStore = store
}

func (wm *WorkflowManager) RegisterTask(name string, taskFunc interface{}) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.tasks[name] = taskFunc
}

func (wm *WorkflowManager) UnregisterTask(name string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.tasks, name)
}

func (wm *WorkflowManager) RegisterHook(name string, hookFunc interface{}) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.hooks[name] = hookFunc
}

func (wm *WorkflowManager) UnregisterHook(name string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.hooks, name)
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
		case "Foreach":
			iterateFunc, ok := wm.tasks[nodeConfig.TaskFunc]
			if !ok {
				return errors.New("iterate function not found: " + nodeConfig.TaskFunc)
			}
			node = &ForeachNode{
				Node: Node[interface{}]{
					ID:   nodeConfig.ID,
					Type: Foreach,
				},
				Collection:  nodeConfig.Collection,
				IterateFunc: wrapTaskFunc(iterateFunc),
			}
		case "Branch":
			branches := make([]NodeInterface, len(nodeConfig.ParallelTasks))
			for i, branchID := range nodeConfig.ParallelTasks {
				branchNode, ok := nodeMap[branchID]
				if !ok {
					return errors.New("branch node not found: " + branchID)
				}
				branches[i] = branchNode
			}
			node = &BranchNode{
				Node: Node[interface{}]{
					ID:   nodeConfig.ID,
					Type: Branch,
				},
				Branches: branches,
			}
		case "Conditional":
			conditionFunc, ok := wm.tasks[nodeConfig.TaskFunc]
			if !ok {
				return errors.New("condition function not found: " + nodeConfig.TaskFunc)
			}
			trueNext, ok := nodeMap[nodeConfig.TrueNext]
			if !ok {
				return errors.New("trueNext node not found: " + nodeConfig.TrueNext)
			}
			falseNext, ok := nodeMap[nodeConfig.FalseNext]
			if !ok {
				return errors.New("falseNext node not found: " + nodeConfig.FalseNext)
			}
			node = &ConditionalNode{
				Node: Node[interface{}]{
					ID:   nodeConfig.ID,
					Type: Conditional,
				},
				Condition: func(data interface{}) bool {
					result, err := wrapTaskFunc(conditionFunc)(data)
					if err != nil {
						return false
					}
					return result.(bool)
				},
				TrueNext:  trueNext,
				FalseNext: falseNext,
			}
		case "SubDag":
			subDag := &Graph{}
			for _, subDagNodeID := range nodeConfig.SubDag {
				subDagNode, ok := nodeMap[subDagNodeID]
				if !ok {
					return errors.New("subDag node not found: " + subDagNodeID)
				}
				subDag.Nodes = append(subDag.Nodes, subDagNode)
			}
			node = &Node[interface{}]{
				ID:     nodeConfig.ID,
				Type:   SubDag,
				SubDag: subDag,
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
		taskFuncValue := reflect.ValueOf(taskFunc)
		if taskFuncValue.Kind() != reflect.Func {
			return nil, fmt.Errorf("taskFunc is not a function")
		}
		if taskFuncValue.Type().NumIn() != 1 || taskFuncValue.Type().NumOut() != 2 {
			return nil, fmt.Errorf("taskFunc should have one input parameter and two output parameters")
		}

		result := taskFuncValue.Call([]reflect.Value{reflect.ValueOf(data)})
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
