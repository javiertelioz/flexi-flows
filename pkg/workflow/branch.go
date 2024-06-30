package workflow

import "fmt"

type BranchNode struct {
	Node
	Branches []NodeInterface
}

func (b *BranchNode) Execute(wm *WorkflowManager, data interface{}) (interface{}, error) {
	fmt.Printf("Executing BranchNode: %s\n", b.ID)
	for _, branch := range b.Branches {
		fmt.Printf("Executing branch with ID: %s\n", branch.GetID())
		if _, err := wm.executeNode(branch, data); err != nil {
			return nil, err
		}
	}
	if len(b.Next) > 0 {
		return wm.executeNode(b.Next[0], data)
	}
	return nil, nil
}
