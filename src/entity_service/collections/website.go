package collections

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/src/pipeline"
	"github.com/omnsight/omndapi/src/utils"
)

func RegisterWebsite(ctx context.Context, client *utils.ArangoDBClient, p *pipeline.Worker) error {
	col, err := client.GetCreateCollection(ctx, "website", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return err
	}
	// Index for URL lookups
	if _, _, err := col.EnsurePersistentIndex(ctx, []string{"url"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_website_url",
		Unique: true,
	}); err != nil {
		return err
	}
	p.RegisterCollection("website", col)
	return nil
}
