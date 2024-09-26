# Go-flows

## Example only code. 
```go
wm := workflow.NewWorkflowManager()
isPrimeNode := &workflow.Node[interface{}]{
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
squareNode := &workflow.Node[any]{
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
    Node: workflow.Node[any]{
        ID:   "parallel",
        Type: workflow.Branch,
    },
    ParallelTasks: []workflow.NodeInterface{isPrimeNode, squareNode},
}

// Nodo para sumar los resultados
sumNode := &workflow.Node[any]{
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

```