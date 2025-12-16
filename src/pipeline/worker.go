package pipeline

import (
	"sync"

	"github.com/arangodb/go-driver"
)

type ConcereteEntityCommon interface {
	GetOwner() string
	GetRead() []string
	GetWrite() []string
}

type Worker struct {
	collections map[string]driver.Collection
	mu          sync.RWMutex
}

func NewWorker() *Worker {
	return &Worker{
		collections: make(map[string]driver.Collection),
	}
}

func (w *Worker) RegisterCollection(entityType string, col driver.Collection) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.collections[entityType] = col
}
