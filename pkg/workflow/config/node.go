package config

type NodeConfig struct {
	ID            string   `json:"id" yaml:"id"`
	Type          string   `json:"type" yaml:"type"`
	TaskFunc      string   `json:"taskFunc,omitempty" yaml:"taskFunc,omitempty"`
	ParallelTasks []string `json:"parallelTasks,omitempty" yaml:"parallelTasks,omitempty"`
	BeforeExecute string   `json:"beforeExecute,omitempty" yaml:"beforeExecute,omitempty"`
	AfterExecute  string   `json:"afterExecute,omitempty" yaml:"afterExecute,omitempty"`
	TrueNext      string   `json:"trueNext,omitempty" yaml:"trueNext,omitempty"`
	FalseNext     string   `json:"falseNext,omitempty" yaml:"falseNext,omitempty"`
	Branches      []string `json:"branches,omitempty" yaml:"branches,omitempty"`
	ExecutionMode string   `json:"executionMode,omitempty" yaml:"executionMode,omitempty"` // Nuevo campo
}
