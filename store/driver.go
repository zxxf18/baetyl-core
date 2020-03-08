package store

type Driver interface {
	Create(key []byte, val []byte) error
	Update(key []byte, res []byte) error
	Delete(key []byte) error
	Get(key []byte) ([]byte, error)
	List(filter func([]byte) bool) ([]byte, error)
	Query(labels map[string]string) ([]byte, error)
	Name() string
}
