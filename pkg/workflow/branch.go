package workflow

import (
	"fmt"
)

type BranchNode struct {
	ID       string
	Type     NodeType
	Branches []NodeInterface
	Next     []NodeInterface
}

func (b *BranchNode) Execute(wm *WorkflowManager) error {
	fmt.Printf("Executing BranchNode: %s\n", b.ID)
	for _, branch := range b.Branches {
		fmt.Printf("Executing branch with ID: %s\n", branch.GetID())
		if err := wm.executeNode(branch); err != nil {
			return err
		}
	}
	if len(b.Next) > 0 {
		return wm.executeNode(b.Next[0])
	}
	return nil
}

func (n *BranchNode) GetID() string {
	return n.ID
}

func (n *BranchNode) GetType() NodeType {
	return n.Type
}
