package workflow

import (
	"errors"
	"sync"
)

type WorkflowManager struct {
	graph *Graph
	mu    sync.Mutex
}

func NewWorkflowManager() *WorkflowManager {
	return &WorkflowManager{
		graph: &Graph{},
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

	_, err := wm.executeNode(startNode, initialData)
	return err
}

func (wm *WorkflowManager) executeNode(node NodeInterface, data interface{}) (interface{}, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}

	result, err := node.Execute(wm, data)
	if err != nil {
		return nil, err
	}

	for _, edge := range wm.graph.Edges {
		if edge.From.GetID() == node.GetID() {
			if edge.To == nil {
				continue
			}
			if edge.Condition == nil || edge.Condition() {
				return wm.executeNode(edge.To, result)
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

func (wm *WorkflowManager) executeSubDag(node *Node, data interface{}) (interface{}, error) {
	subDag := node.SubDag
	var result interface{}
	var err error
	for _, subNode := range subDag.Nodes {
		result, err = wm.executeNode(subNode, data)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (wm *WorkflowManager) executeConditional(node *Node, data interface{}) (interface{}, error) {
	for _, edge := range wm.graph.Edges {
		if edge.From.GetID() == node.GetID() && edge.Condition() {
			return wm.executeNode(edge.To, data)
		}
	}
	return nil, nil
}
