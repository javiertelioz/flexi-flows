package main

import (
	"fmt"
	"log"

	"github.com/javiertelioz/workflows/pkg/workflow"
)

func main() {
	// Crear una instancia de WorkflowManager
	wm := workflow.NewWorkflowManager()

	// Crear nodos para el sub-dag
	subTaskNode1 := &workflow.Node{
		ID:   "subtask1",
		Type: workflow.Task,
		TaskFunc: func() error {
			fmt.Println("Executing SubTask 1")
			return nil
		},
	}

	subTaskNode2 := &workflow.Node{
		ID:   "subtask2",
		Type: workflow.Task,
		TaskFunc: func() error {
			fmt.Println("Executing SubTask 2")
			return nil
		},
	}

	// Crear sub-dag
	subDag := &workflow.Graph{
		Nodes: []workflow.NodeInterface{subTaskNode1, subTaskNode2},
		Edges: []*workflow.Edge{
			{
				From: subTaskNode1,
				To:   subTaskNode2,
			},
		},
	}

	// Crear nodo de tipo SubDag
	subDagNode := &workflow.Node{
		ID:     "subdag",
		Type:   workflow.SubDag,
		SubDag: subDag,
	}

	// Crear nodos principales
	taskNode1 := &workflow.Node{
		ID:   "task1",
		Type: workflow.Task,
		TaskFunc: func() error {
			fmt.Println("Executing Task 1")
			return nil
		},
	}

	taskNode2 := &workflow.Node{
		ID:   "task2",
		Type: workflow.Task,
		TaskFunc: func() error {
			fmt.Println("Executing Task 2")
			return nil
		},
	}

	foreachNode := &workflow.ForeachNode{
		ID:         "foreach1",
		Type:       workflow.Foreach,
		Collection: []interface{}{1, 2, 3},
		IterateFunc: func(item interface{}) error {
			fmt.Printf("Processing item: %v\n", item)
			return nil
		},
	}

	branchNode := &workflow.BranchNode{
		ID:       "branch1",
		Type:     workflow.Branch,
		Branches: []workflow.NodeInterface{foreachNode},
	}

	wm.AddNode(taskNode1)
	wm.AddNode(taskNode2)
	wm.AddNode(foreachNode)
	wm.AddNode(branchNode)
	wm.AddNode(subDagNode)

	wm.AddEdge(&workflow.Edge{
		From: taskNode1,
		To:   branchNode,
	})
	wm.AddEdge(&workflow.Edge{
		From: branchNode,
		To:   foreachNode,
	})
	wm.AddEdge(&workflow.Edge{
		From: foreachNode,
		To:   subDagNode,
	})
	wm.AddEdge(&workflow.Edge{
		From: subDagNode,
		To:   taskNode2,
	})
	wm.AddEdge(&workflow.Edge{
		From: taskNode2,
		To:   nil,
	})

	err := wm.Execute("task1")
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
