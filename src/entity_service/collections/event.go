package collections

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/src/pipeline"
	"github.com/omnsight/omndapi/src/utils"
)

func RegisterEvent(ctx context.Context, client *utils.ArangoDBClient, p *pipeline.Worker) error {
	col, err := client.GetCreateCollection(ctx, "event", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return err
	}
	// Index for time range queries
	if _, _, err := col.EnsurePersistentIndex(ctx, []string{"happened_at"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_event_happened_at",
	}); err != nil {
		return err
	}
	p.RegisterCollection("event", col)
	return nil
}
