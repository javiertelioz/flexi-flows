package storage

type StateStore[T any] interface {
	SaveState(nodeID string, data T) error
	LoadState(nodeID string) (T, bool, error)
}
