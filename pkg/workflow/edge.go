package workflow

type Edge[T any] struct {
	From      NodeInterface[T]
	To        NodeInterface[T]
	Condition func() bool
}
