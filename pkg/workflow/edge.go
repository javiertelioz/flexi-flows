package workflow

type Edge struct {
	From      NodeInterface
	To        NodeInterface
	Condition func() bool
}
