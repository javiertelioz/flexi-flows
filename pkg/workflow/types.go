package workflow

type NodeType int

const (
	Task NodeType = iota
	SubDag
	Conditional
	Foreach
	Branch
	Parallel
)
