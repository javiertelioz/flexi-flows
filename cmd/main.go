// main.go
package main

import (
	"fmt"
	"github.com/javiertelioz/workflows/pkg/workflow"
	"log"
	"math"
)

func main() {
	// Crear una instancia de StateStore
	// stateStore := workflow.NewJSONStateStore("flow.json")

	// Crear una instancia de WorkflowManager con StateStore
	wm := workflow.NewWorkflowManager(nil)

	// Nodo para verificar si un número es primo
	isPrimeNode := &workflow.Node{
		ID:   "isPrime",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			num := data.(int)
			if num <= 1 {
				return false, nil
			}
			for i := 2; i <= int(math.Sqrt(float64(num))); i++ {
				if num%i == 0 {
					return false, nil
				}
			}
			return true, nil
		},
		BeforeExecute: func(data interface{}) (interface{}, error) {
			fmt.Println("Before isPrime: received data", data)
			return data, nil
		},
		AfterExecute: func(result interface{}) (interface{}, error) {
			fmt.Println("After isPrime: result is", result)
			return result, nil
		},
	}

	// Nodo para multiplicar un número por sí mismo
	squareNode := &workflow.Node{
		ID:   "square",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			num := data.(int)
			return num * num, nil
		},
		BeforeExecute: func(data interface{}) (interface{}, error) {
			fmt.Println("Before square: received data", data)
			return data, nil
		},
		AfterExecute: func(result interface{}) (interface{}, error) {
			fmt.Println("After square: result is", result)
			return result, nil
		},
	}

	// Nodo paralelo para ejecutar las dos tareas anteriores simultáneamente
	parallelNode := &workflow.ParallelNode{
		Node: workflow.Node{
			ID:   "parallel",
			Type: workflow.Branch,
		},
		ParallelTasks: []workflow.NodeInterface{isPrimeNode, squareNode},
	}

	// Nodo para sumar los resultados
	sumNode := &workflow.Node{
		ID:   "sum",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			results := data.([]interface{})
			isPrime := results[0].(bool)
			square := results[1].(int)
			sum := square
			if isPrime {
				sum += 1
			}
			fmt.Printf("Sum of results: %d (Prime: %v, Square: %d)\n", sum, isPrime, square)
			return sum, nil
		},
		BeforeExecute: func(data interface{}) (interface{}, error) {
			fmt.Println("Before sum: received data", data)
			return data, nil
		},
		AfterExecute: func(result interface{}) (interface{}, error) {
			fmt.Println("After sum: result is", result)
			return result, nil
		},
	}

	wm.AddNode(isPrimeNode)
	wm.AddNode(squareNode)
	wm.AddNode(parallelNode)
	wm.AddNode(sumNode)

	wm.AddEdge(&workflow.Edge{
		From: parallelNode,
		To:   sumNode,
	})

	err := wm.Execute("parallel", 5)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
	/*// Crear una instancia de WorkflowManager
	wm := workflow.NewWorkflowManager()

	// Crear nodos para el sub-dag
	subTaskNode1 := &workflow.Node{
		ID:   "subtask1",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			fmt.Println("Executing SubTask 1")
			return "result from subtask 1", nil
		},
		BeforeExecute: func() {
			fmt.Println("Starting SubTask 1")
		},
		AfterExecute: func() {
			fmt.Println("Finished SubTask 1")
		},
	}

	subTaskNode2 := &workflow.Node{
		ID:   "subtask2",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			fmt.Println("Executing SubTask 2 with data:", data)
			return nil, nil
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
		TaskFunc: func(data interface{}) (interface{}, error) {
			fmt.Println("Executing Task 1")
			return "result from task 1", nil
		},
		AfterExecute: func() {
			fmt.Println("Finished Task 1")
		},
	}

	taskNode2 := &workflow.Node{
		ID:   "task2",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			fmt.Println("Executing Task 2 with data:", data)
			return nil, nil
		},
		BeforeExecute: func() {
			fmt.Println("Starting task 2")
		},
	}

	foreachNode := &workflow.ForeachNode{
		Node: workflow.Node{
			ID:   "foreach1",
			Type: workflow.Foreach,
		},
		Collection: []interface{}{1, 2, 3},
		IterateFunc: func(item interface{}) (interface{}, error) {
			fmt.Printf("Processing item: %v\n", item)
			return nil, nil
		},
	}

	branchNode := &workflow.BranchNode{
		Node: workflow.Node{
			ID:   "branch1",
			Type: workflow.Branch,
		},
		Branches: []workflow.NodeInterface{foreachNode},
	}

	conditionalNode := &workflow.ConditionalNode{
		Node: workflow.Node{
			ID:   "conditional1",
			Type: workflow.Conditional,
		},
		Condition: func(data interface{}) bool {
			result, ok := data.(string)
			return ok && result == "result from task 1"
		},
		TrueNext:  subDagNode,
		FalseNext: taskNode2,
	}

	wm.AddNode(taskNode1)
	wm.AddNode(taskNode2)
	wm.AddNode(foreachNode)
	wm.AddNode(branchNode)
	wm.AddNode(subDagNode)
	wm.AddNode(conditionalNode)

	wm.AddEdge(&workflow.Edge{
		From: taskNode1,
		To:   conditionalNode,
	})
	wm.AddEdge(&workflow.Edge{
		From: conditionalNode,
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

	err := wm.Execute("task1", nil)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}*/
}
