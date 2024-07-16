package main

import "github.com/javiertelioz/go-flows/examples/subdag_flow"

func main() {
	//http_flow.HttpFlow()
	//task_flow.TaskFlow()
	//conditional_flow.ConditionalFlow()
	//foreach_flow.ForeachFlow()
	//branch_flow.BranchFlow()
	subdag_flow.SubDagFlow()
	//number_flow.NumberFlow()

	/*srcDir := "./examples"
	metadata, err := comment.ParseComments(srcDir)
	if err != nil {
		log.Fatalf("Failed to parse comments: %v", err)
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metadata to JSON: %v", err)
	}

	fmt.Printf("Function metadata:\n%s\n", metadataJSON)*/

}
