package entityservice

import (
	"context"

	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/entity_service/collections"
	"github.com/omnsight/omndapi/src/pipeline"
	"github.com/omnsight/omndapi/src/utils"
)

type EntityService struct {
	dapi.UnimplementedEntityServiceServer

	DBClient *utils.ArangoDBClient
	Pipeline *pipeline.Worker
}

func NewEntityService(client *utils.ArangoDBClient) (*EntityService, error) {
	service := &EntityService{
		DBClient: client,
		Pipeline: pipeline.NewWorker(),
	}

	ctx := context.Background()

	// Register collections with specific indices
	// Execute registrations
	if err := collections.RegisterEvent(ctx, client, service.Pipeline); err != nil {
		return nil, err
	}
	if err := collections.RegisterSource(ctx, client, service.Pipeline); err != nil {
		return nil, err
	}
	if err := collections.RegisterWebsite(ctx, client, service.Pipeline); err != nil {
		return nil, err
	}
	if err := collections.RegisterPerson(ctx, client, service.Pipeline); err != nil {
		return nil, err
	}
	if err := collections.RegisterOrganization(ctx, client, service.Pipeline); err != nil {
		return nil, err
	}

	return service, nil
}
