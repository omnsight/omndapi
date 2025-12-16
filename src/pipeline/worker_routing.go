package pipeline

import (
	"fmt"

	"github.com/arangodb/go-driver"
)

// GetCollection retrieves the registered collection for the given entity type.
func (w *Worker) GetCollection(entityType string) (driver.Collection, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	col, ok := w.collections[entityType]
	if !ok {
		return nil, fmt.Errorf("collection for entity type '%s' not found", entityType)
	}
	return col, nil
}