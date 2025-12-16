package relationshipservice

import (
	"context"
	"encoding/json"

	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RelationshipService) UpdateRelationship(ctx context.Context, req *dapi.UpdateRelationshipRequest) (*dapi.UpdateRelationshipResponse, error) {
	userId, userRoles, err := s.Pipeline.GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update relationship: %s/%s", userId, userRoles, req.GetCollection(), req.GetKey())

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

	if err := s.Pipeline.CheckWritePermission(&existingRelationship, userId, userRoles); err != nil {
		return nil, err
	}

	relationshipToUpdate := req.GetRelationship()
	if err := s.Pipeline.SetPermissions(relationshipToUpdate, userId, false); err != nil {
		return nil, err
	}

	data, err := json.Marshal(relationshipToUpdate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal update data")
	}
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal update data")
	}

	delete(dataMap, "_id")
	delete(dataMap, "_key")
	delete(dataMap, "_rev")

	var updatedRelationship model.Relation
	meta, err := s.Pipeline.UpdateDocument(ctx, col, req.GetKey(), dataMap, &updatedRelationship)
	if err != nil {
		return nil, err
	}

	updatedRelationship.Id = meta.ID.String()
	updatedRelationship.Key = meta.Key
	updatedRelationship.Rev = meta.Rev
	return &dapi.UpdateRelationshipResponse{Relationship: &updatedRelationship}, nil
}
