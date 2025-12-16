package collections

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/src/pipeline"
	"github.com/omnsight/omndapi/src/utils"
)

func RegisterOrganization(ctx context.Context, client *utils.ArangoDBClient, p *pipeline.Worker) error {
	col, err := client.GetCreateCollection(ctx, "organization", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return err
	}
	// Index for organization names
	if _, _, err := col.EnsurePersistentIndex(ctx, []string{"name"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_organization_name",
	}); err != nil {
		return err
	}
	p.RegisterCollection("organization", col)
	return nil
}
