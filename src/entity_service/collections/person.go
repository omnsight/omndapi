package collections

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/src/pipeline"
	"github.com/omnsight/omndapi/src/utils"
)

func RegisterPerson(ctx context.Context, client *utils.ArangoDBClient, p *pipeline.Worker) error {
	col, err := client.GetCreateCollection(ctx, "person", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return err
	}
	// Index for person names
	if _, _, err := col.EnsurePersistentIndex(ctx, []string{"name"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_person_name",
	}); err != nil {
		return err
	}
	p.RegisterCollection("person", col)
	return nil
}
