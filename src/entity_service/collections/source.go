package collections

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/src/pipeline"
	"github.com/omnsight/omndapi/src/utils"
)

func RegisterSource(ctx context.Context, client *utils.ArangoDBClient, p *pipeline.Worker) error {
	col, err := client.GetCreateCollection(ctx, "source", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return err
	}
	// Index for source name if needed
	if _, _, err := col.EnsurePersistentIndex(ctx, []string{"url"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_source_url",
	}); err != nil {
		return err
	}
	p.RegisterCollection("source", col)
	return nil
}
