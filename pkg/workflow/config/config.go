package config

type NodeConfig struct {
	ID            string   `json:"id" yaml:"id"`
	Type          string   `json:"type" yaml:"type"`
	TaskFunc      string   `json:"taskFunc,omitempty" yaml:"taskFunc,omitempty"`
	ParallelTasks []string `json:"parallelTasks,omitempty" yaml:"parallelTasks,omitempty"`
	BeforeExecute string   `json:"beforeExecute,omitempty" yaml:"beforeExecute,omitempty"`
	AfterExecute  string   `json:"afterExecute,omitempty" yaml:"afterExecute,omitempty"`
}

type EdgeConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type WorkflowConfig struct {
	Nodes []NodeConfig `json:"nodes"`
	Edges []EdgeConfig `json:"edges"`
}
