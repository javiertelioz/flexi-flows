package workflow

type Graph[T any] struct {
	Nodes []NodeInterface[T]
	Edges []*Edge[T]
}
