package workflow

type StateStore interface {
	SaveState(nodeID string, data interface{}) error
	LoadState(nodeID string) (interface{}, error)
}
