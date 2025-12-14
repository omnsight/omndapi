package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RelationshipService struct {
	dapi.UnimplementedRelationshipServiceServer

	DBClient *utils.ArangoDBClient
}

func NewRelationshipService(client *utils.ArangoDBClient) (*RelationshipService, error) {
	service := &RelationshipService{
		DBClient: client,
	}

	return service, nil
}

func (s *RelationshipService) CreateRelationship(ctx context.Context, req *dapi.CreateRelationshipRequest) (*dapi.CreateRelationshipResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.WithFields(logrus.Fields{
		"from": req.GetRelationship().GetFrom(),
		"to":   req.GetRelationship().GetTo(),
	}).Infof("[%s, %v] requests to create relationship", userId, userRoles)

	if !slices.Contains(userRoles, "admin") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	relationship := req.GetRelationship()
	if relationship == nil {
		logger.Error("relationship is nil")
		return nil, status.Errorf(codes.InvalidArgument, "Bad parameter")
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

func (s *RelationshipService) UpdateRelationship(ctx context.Context, req *dapi.UpdateRelationshipRequest) (*dapi.UpdateRelationshipResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update relationship: %s/%s", userId, userRoles, req.GetCollection(), req.GetKey())

	if !slices.Contains(userRoles, "admin") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	// Using AQL query to update the document with arangodb ID
	query := `
		LET cleanPatch = UNSET(@patch, "_id", "_key", "_rev")
		UPDATE @key WITH cleanPatch IN @@collection
		RETURN NEW
	`

	cursor, err := s.DBClient.DB.Query(ctx, query, map[string]interface{}{
		"key":         req.GetKey(),
		"patch":       req.GetRelationship(),
		"@collection": req.GetCollection(),
	})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"data":  req.GetRelationship(),
		}).Error("failed to execute AQL query for updating relationship")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}
	defer cursor.Close()

	var relationship model.Relation
	meta, err := cursor.ReadDocument(ctx, &relationship)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			logger.WithFields(logrus.Fields{
				"collection": req.GetCollection(),
				"key":        req.GetKey(),
			}).Info("relationship not found for update")
			return nil, status.Errorf(codes.NotFound, "Relation not found")
		}

		logger.WithFields(logrus.Fields{
			"error":      err,
			"collection": req.GetCollection(),
			"key":        req.GetKey(),
		}).Error("failed to read updated relationship document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	relationship.Id = meta.ID.String()
	relationship.Key = meta.Key
	relationship.Rev = meta.Rev
	return &dapi.UpdateRelationshipResponse{Relationship: &relationship}, nil
}

func (s *RelationshipService) DeleteRelationship(ctx context.Context, req *dapi.DeleteRelationshipRequest) (*dapi.DeleteRelationshipResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to delete relationship: %s/%s", userId, userRoles, req.GetCollection(), req.GetKey())

	if !slices.Contains(userRoles, "admin") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	// Using AQL query to delete the document with arangodb ID
	query := `
		FOR doc IN @@collection
			FILTER doc._key == @key
			REMOVE doc IN @@collection
			RETURN OLD
	`

	cursor, err := s.DBClient.DB.Query(ctx, query, map[string]interface{}{
		"key":         req.GetKey(),
		"@collection": req.GetCollection(),
	})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error":      err,
			"collection": req.GetCollection(),
			"key":        req.GetKey(),
		}).Error("failed to execute AQL query for deleting relationship")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}
	defer cursor.Close()

	var relationship model.Relation
	_, err = cursor.ReadDocument(ctx, &relationship)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			logger.WithFields(logrus.Fields{
				"collection": req.GetCollection(),
				"key":        req.GetKey(),
			}).Info("relationship not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Relation not found")
		}

		logger.WithFields(logrus.Fields{
			"error":      err,
			"collection": req.GetCollection(),
			"key":        req.GetKey(),
		}).Error("failed to read deleted relationship document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.DeleteRelationshipResponse{}, nil
}
