package relationshipservice

import (
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/pipeline"
	"github.com/omnsight/omndapi/src/utils"
)

type RelationshipService struct {
	dapi.UnimplementedRelationshipServiceServer

	DBClient *utils.ArangoDBClient
	Pipeline *pipeline.Worker
}

func NewRelationshipService(client *utils.ArangoDBClient) (*RelationshipService, error) {
	service := &RelationshipService{
		DBClient: client,
		Pipeline: pipeline.NewWorker(),
	}

	return service, nil
}
