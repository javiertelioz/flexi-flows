package nodes

import (
	"fmt"
	"github.com/javiertelioz/go-flows/pkg/workflow"
)

type BranchNode struct {
	workflow.Node[interface{}]
	Branches []workflow.NodeInterface
}

func (b *BranchNode) Execute(wm *workflow.WorkflowManager, data interface{}) (interface{}, error) {
	fmt.Printf("Executing BranchNode: %s\n", b.ID)
	for _, branch := range b.Branches {
		fmt.Printf("Executing branch with ID: %s\n", branch.GetID())
		if _, err := wm.ExecuteNode(branch, data); err != nil {
			return nil, err
		}
	}
	if len(b.Next) > 0 {
		return wm.ExecuteNode(b.Next[0], data)
	}
	return nil, nil
}
