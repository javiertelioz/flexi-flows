package config

type WorkflowConfig struct {
	Nodes []NodeConfig `json:"nodes" yaml:"nodes"`
	Edges []EdgeConfig `json:"edges" yaml:"edges"`
}

type EdgeConfig struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}
