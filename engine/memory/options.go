package memory

import "github.com/rafalb8/go-storage"

type MemoryOpts func(*InMemory) error

func Logger(lg storage.Logger) MemoryOpts {
	return func(m *InMemory) error {
		m.lg = lg
		return nil
	}
}
