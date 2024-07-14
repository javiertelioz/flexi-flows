package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/javiertelioz/go-flows/pkg/workflow/comment"
	"log"
	"math"
	"net/http"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/javiertelioz/go-flows/pkg/workflow/config"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// getUserFunc
// @type: task
// @description Retrieve user information by id.
//
// @input userID (int): User ID.
//
// @output User: Returns user information.
// @output error: Returns an error if user no exist.
func getUserFunc(userID int) (*User, error) {
	if userID == 0 {
		return nil, errors.New("error to retrieve user")
	}

	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%d", userID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	fmt.Printf("User ID: %d\nUser Name: %s\n", user.ID, user.Name)
	return &user, nil

}

// isPrimeFunc
// @type: task
// @description checks if a given integer is a prime number.
//
// @input data (int): The integer to be checked for primality.
//
// @output bool: Returns true if the input integer is a prime number, false otherwise.
// @output error: Returns an error if the input integer is less than 1.
func isPrimeFunc(data int) (bool, error) {
	if data <= 1 {
		return false, nil
	}
	for i := 2; i <= int(math.Sqrt(float64(data))); i++ {
		if data%i == 0 {
			return false, nil
		}
	}
	return true, nil
}

// beforeIsPrime
// @type: pre-hook
// @before isPrimeFunc
// @description execute before isPrimeFunc.
//
// @input data (int): The integer to be checked for primality.
//
// @output int: Returns integer is a prime number.
// @output error: Returns an error if occurs.
func beforeIsPrime(data int) (int, error) {
	fmt.Println("Before isPrime: received data", data)
	return data, nil
}

// afterIsPrime
// @type: post-hook
// @after isPrimeFunc
// @description execute after isPrimeFunc.
//
// @input result (bool): The result of the isPrimeFunc.
//
// @output bool: Returns the result after post-processing.
// @output error: Returns an error if occurs.
func afterIsPrime(result bool) (bool, error) {
	fmt.Println("After isPrime: result is", result)
	return result, nil
}

func squareFunc(data int) (int, error) {
	return data * data, nil
}

func sumFunc(data []interface{}) (int, error) {
	isPrime := data[0].(bool)
	square := data[1].(int)
	sum := square
	if isPrime {
		sum += 1
	}
	fmt.Printf("Sum of results: %d (Prime: %v, Square: %d)\n", sum, isPrime, square)
	return sum, nil
}

func beforeSquare(data int) (int, error) {
	fmt.Println("Before square: received data", data)
	return data, nil
}

func afterSquare(result int) (int, error) {
	fmt.Println("After square: result is", result)
	return result, nil
}

func beforeSum(data []interface{}) ([]interface{}, error) {
	fmt.Println("Before sum: received data", data)
	return data, nil
}

func afterSum(result int) (int, error) {
	fmt.Println("After sum: result is", result)
	return result, nil
}

func main() {
	// Crear una instancia de StateStore (opcional)
	//stateStore := storage.NewJSONStateStore("flow.json")

	wm := workflow.NewWorkflowManager()

	// Register storage
	//wm.RegisterStateStore(stateStore)

	// Register task & hooks
	wm.RegisterTask("getUserFunc", getUserFunc)
	wm.RegisterTask("isPrimeFunc", isPrimeFunc)
	wm.RegisterTask("squareFunc", squareFunc)
	wm.RegisterTask("sumFunc", sumFunc)

	wm.RegisterHook("beforeIsPrime", beforeIsPrime)
	wm.RegisterHook("afterIsPrime", afterIsPrime)
	wm.RegisterHook("beforeSquare", beforeSquare)
	wm.RegisterHook("afterSquare", afterSquare)
	wm.RegisterHook("beforeSum", beforeSum)
	wm.RegisterHook("afterSum", afterSum)

	// Cargar la configuración del flujo de trabajo desde un archivo JSON
	workflowConfig, err := config.LoadConfig("config/workflow.yaml")
	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	// Crear nodos y edges basados en la configuración
	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	// Ejecutar el flujo de trabajo
	err = wm.Execute("parallel", 5)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	srcDir := "./cmd"
	metadata, err := comment.ParseComments(srcDir)
	if err != nil {
		log.Fatalf("Failed to parse comments: %v", err)
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metadata to JSON: %v", err)
	}

	fmt.Printf("Function metadata:\n%s\n", metadataJSON)

	// Nodo para verificar si un número es primo
	/* isPrimeNode := &workflow.Node{
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
	}*/
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
