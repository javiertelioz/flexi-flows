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

func (wm *WorkflowManager) Execute(startNodeID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	startNode := wm.findNodeByID(startNodeID)
	if startNode == nil {
		return errors.New("start node not found")
	}

	return wm.executeNode(startNode)
}

func (wm *WorkflowManager) executeNode(node NodeInterface) error {
	if node == nil {
		return errors.New("node is nil")
	}

	switch node.GetType() {
	case Task:
		n := node.(*Node)
		if err := n.Execute(wm); err != nil {
			return err
		}
	case SubDag:
		n := node.(*Node)
		return wm.executeSubDag(n)
	case Conditional:
		n := node.(*Node)
		return wm.executeConditional(n)
	case Foreach, Branch:
		if err := node.Execute(wm); err != nil {
			return err
		}
	}

	for _, edge := range wm.graph.Edges {
		if edge.From.GetID() == node.GetID() {
			if edge.To == nil {
				continue
			}
			if edge.Condition == nil || edge.Condition() {
				if err := wm.executeNode(edge.To); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (wm *WorkflowManager) findNodeByID(id string) NodeInterface {
	for _, node := range wm.graph.Nodes {
		if node.GetID() == id {
			return node
		}
	}
	return nil
}

func (wm *WorkflowManager) executeSubDag(node *Node) error {
	subDag := node.SubDag
	for _, subNode := range subDag.Nodes {
		if err := wm.executeNode(subNode); err != nil {
			return err
		}
	}
	return nil
}

func (wm *WorkflowManager) executeConditional(node *Node) error {
	for _, edge := range wm.graph.Edges {
		if edge.From.GetID() == node.GetID() && edge.Condition() {
			return wm.executeNode(edge.To)
		}
	}
	return nil
}
