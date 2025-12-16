package relationshipservice

import (
	"context"

	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RelationshipService) DeleteRelationship(ctx context.Context, req *dapi.DeleteRelationshipRequest) (*dapi.DeleteRelationshipResponse, error) {
	userId, userRoles, err := s.Pipeline.GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to delete relationship: %s/%s", userId, userRoles, req.GetCollection(), req.GetKey())

	col, err := s.DBClient.DB.Collection(ctx, req.GetCollection())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error":      err,
			"collection": req.GetCollection(),
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.NotFound, "Collection not found")
	}

	var existingRelationship model.Relation
	if _, err := s.Pipeline.ReadDocument(ctx, col, req.GetKey(), &existingRelationship); err != nil {
		return nil, err
	}

	if err := s.Pipeline.CheckDeletePermission(&existingRelationship, userId); err != nil {
		return nil, err
	}

	if err := s.Pipeline.DeleteDocument(ctx, col, req.GetKey()); err != nil {
		return nil, err
	}

	return &dapi.DeleteRelationshipResponse{}, nil
}
