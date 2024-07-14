package config

type WorkflowConfig struct {
	Nodes []NodeConfig `json:"nodes"`
	Edges []EdgeConfig `json:"edges"`
}

type EdgeConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
}
