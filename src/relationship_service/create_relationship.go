package relationshipservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RelationshipService) CreateRelationship(ctx context.Context, req *dapi.CreateRelationshipRequest) (*dapi.CreateRelationshipResponse, error) {
	userId, userRoles, err := s.Pipeline.GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.WithFields(logrus.Fields{
		"from": req.GetRelationship().GetFrom(),
		"to":   req.GetRelationship().GetTo(),
	}).Infof("[%s, %v] requests to create relationship", userId, userRoles)

	if err := s.Pipeline.CheckCreatePermission(userRoles); err != nil {
		return nil, err
	}

	relationship := req.GetRelationship()
	if relationship == nil {
		logger.Error("relationship is nil")
		return nil, status.Errorf(codes.InvalidArgument, "Bad parameter")
	}

	if err := s.Pipeline.SetPermissions(relationship, userId, true); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to set permissions")
		return nil, err
	}

	fromColl, _, err := s.DBClient.ParseDocID(relationship.From)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    relationship.From,
		}).Error("failed to parse from entitity id")
		return nil, status.Errorf(codes.InvalidArgument, "Bad parameter")
	}

	toColl, _, err := s.DBClient.ParseDocID(relationship.To)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    relationship.From,
		}).Error("failed to parse to entitity id")
		return nil, status.Errorf(codes.InvalidArgument, "Bad parameter")
	}

	// Process relation name
	relationName := strings.ToLower(strings.ReplaceAll(relationship.Name, " ", "_"))
	if len(relationName) == 0 {
		logger.Error("invalid relation name")
		return nil, status.Errorf(codes.InvalidArgument, "invalid relation name")
	}

	collectionName := fmt.Sprintf("%s_%s_%s", fromColl, relationName, toColl)

	// Create the edge collection if it doesn't exist
	collection, err := s.DBClient.GetCreateEdgeCollection(ctx, collectionName, driver.VertexConstraints{
		From: []string{fromColl},
		To:   []string{toColl},
	}, driver.CreateEdgeCollectionOptions{})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"name":  collectionName,
		}).Errorf("failed to get or create collection %s", collectionName)
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Create document in collection
	relationship.Id = ""
	relationship.Key = ""
	relationship.Rev = ""

	var createdRelationship model.Relation
	ctxWithReturnNew := driver.WithReturnNew(ctx, &createdRelationship)
	meta, err := collection.CreateDocument(ctxWithReturnNew, relationship)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"data":  relationship,
		}).Error("failed to create relationship document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	createdRelationship.Id = meta.ID.String()
	createdRelationship.Key = meta.Key
	createdRelationship.Rev = meta.Rev
	return &dapi.CreateRelationshipResponse{Relationship: &createdRelationship}, nil
}
