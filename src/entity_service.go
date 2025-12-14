package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EntityService struct {
	dapi.UnimplementedEntityServiceServer

	DBClient    *utils.ArangoDBClient
	Collections map[string]driver.Collection
	mu          sync.RWMutex
}

func NewEntityService(client *utils.ArangoDBClient) (*EntityService, error) {
	service := &EntityService{
		DBClient:    client,
		Collections: make(map[string]driver.Collection),
	}
	return service, nil
}

func (s *EntityService) ListEntitiesFromEvent(ctx context.Context, req *dapi.ListEntitiesFromEventRequest) (*dapi.ListEntitiesFromEventResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to list entities from event", userId, userRoles)

	// AQL query to find start events, filter them, traverse, and get relations
	query := `
		LET start_events = (
			@startNode != "" ? (
				FOR e IN event FILTER e._id == @startNode RETURN e
			) : (
				FOR e IN event
				FILTER e.happened_at >= @startTime AND e.happened_at <= @endTime
				RETURN e
			)
		)

		LET filtered_events = (
			FOR e IN start_events
			FILTER (@countryCode == "" OR e.location.countryCode == @countryCode OR e.location.country_code == @countryCode)
			FILTER (@tag == "" OR @tag IN e.tags)
			RETURN e
		)

		// Traversal to find all connected entities within depth
		LET traversed_nodes = (
			FOR start_node IN filtered_events
				FOR v IN 0..@depth ANY start_node GRAPH @graphName
				OPTIONS {uniqueVertices: 'global', bfs: true}
				RETURN DISTINCT v
		)

		// Find relations among all found entities
		// We look for edges where both 'from' and 'to' are in our set of traversed nodes
		LET relations = (
			FOR id IN traversed_nodes[*]._id
				FOR v, e IN 1..1 ANY id GRAPH @graphName
				FILTER v._id IN traversed_nodes[*]._id
				RETURN DISTINCT e
		)

		// Return typed entities and relations
		RETURN { 
			entities: (
				FOR doc IN traversed_nodes
				RETURN { type: PARSE_IDENTIFIER(doc._id).collection, data: doc }
			), 
			relations: relations 
		}
	`

	bindVars := map[string]interface{}{
		"startNode":   req.GetStartNode(),
		"startTime":   req.GetStartTime(),
		"endTime":     req.GetEndTime(),
		"countryCode": req.GetCountryCode(),
		"tag":         req.GetTag(),
		"depth":       req.GetDepth(),
		"graphName":   s.DBClient.OsintGraph.Name(),
	}

	cursor, err := s.DBClient.DB.Query(ctx, query, bindVars)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"query": query,
			"vars":  bindVars,
		}).Error("failed to execute AQL query")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}
	defer cursor.Close()

	type EntityResult struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	type QueryResult struct {
		Entities  []EntityResult   `json:"entities"`
		Relations []model.Relation `json:"relations"`
	}

	var result QueryResult
	_, err = cursor.ReadDocument(ctx, &result)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			return &dapi.ListEntitiesFromEventResponse{}, nil
		}
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to read query result")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	var pbEntities []*model.Entity
	for _, er := range result.Entities {
		entityObj, err := utils.GetEmptyStruct(er.Type)
		if err != nil {
			logger.Warnf("unknown entity type %s: %v", er.Type, err)
			continue
		}

		if err = json.Unmarshal(er.Data, entityObj); err != nil {
			logger.Errorf("failed to unmarshal entity data for type %s: %v", er.Type, err)
			continue
		}

		entity, err := utils.WrapEntity(er.Type, entityObj)
		if err != nil {
			logger.Errorf("failed to wrap entity %s: %v", er.Type, err)
			continue
		}
		pbEntities = append(pbEntities, entity)
	}

	var pbRelations []*model.Relation
	for i := range result.Relations {
		pbRelations = append(pbRelations, &result.Relations[i])
	}

	return &dapi.ListEntitiesFromEventResponse{
		Entities:  pbEntities,
		Relations: pbRelations,
	}, nil
}

func (s *EntityService) GetEntity(ctx context.Context, req *dapi.GetEntityRequest) (*dapi.GetEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to get %s with ID: %s", userId, userRoles, req.GetEntityType(), req.GetKey())

	collection, err := s.getCollection(ctx, req.GetEntityType())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	entityObj, err := utils.GetEmptyStruct(req.GetEntityType())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to parse empty struct")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	meta, err := collection.ReadDocument(ctx, req.GetKey(), entityObj)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"entity_type": req.GetEntityType(),
				"key":         req.GetKey(),
			}).Info("entity not found")
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to read entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if err = utils.SetEntityMeta(req.GetEntityType(), entityObj, meta); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to set entity meta")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	entityRoles, err := utils.GetEntityRoles(req.GetEntityType(), entityObj)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to get entity roles")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if len(lo.Intersect(entityRoles, userRoles)) == 0 && !slices.Contains(userRoles, "admin") {
		logger.WithFields(logrus.Fields{
			"key":          req.GetKey(),
			"user":         userId,
			"user_roles":   userRoles,
			"entity_roles": entityRoles,
		}).Info("User does not have permission to access this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required roles")
	}

	entity, err := utils.WrapEntity(req.GetEntityType(), entityObj)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to wrap entity")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.GetEntityResponse{
		Entity: entity,
	}, nil
}

func (s *EntityService) CreateEntity(ctx context.Context, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to create %s", userId, userRoles, req.GetEntityType())

	if !slices.Contains(userRoles, "admin") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	typeStr, entityObj, err := utils.UnwrapEntity(req.GetEntity())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error":  err,
			"entity": entityObj,
		}).Error("failed to unwrap entity")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if typeStr != req.GetEntityType() {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"parsed_type": typeStr,
		}).Error("entity type mismatch")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	collection, err := s.getCollection(ctx, req.GetEntityType())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	createdEntity, err := utils.GetEmptyStruct(req.GetEntityType())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get empty struct for result")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	ctxWithReturnNew := driver.WithReturnNew(ctx, createdEntity)
	meta, err := collection.CreateDocument(ctxWithReturnNew, entityObj)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to create entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if err = utils.SetEntityMeta(req.GetEntityType(), createdEntity, meta); err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to set entity meta")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	entity, err := utils.WrapEntity(req.GetEntityType(), createdEntity)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to wrap entity")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// We return the same entity as created
	return &dapi.CreateEntityResponse{
		Entity: entity,
	}, nil
}

func (s *EntityService) UpdateEntity(ctx context.Context, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update %s with ID: %s", userId, userRoles, req.GetEntityType(), req.GetKey())

	if !slices.Contains(userRoles, "admin") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	typeStr, entityObj, err := utils.UnwrapEntity(req.GetEntity())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to unwrap entity")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if typeStr != req.GetEntityType() {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"parsed_type": typeStr,
		}).Error("entity type mismatch")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	collection, err := s.getCollection(ctx, req.GetEntityType())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	updatedEntity, err := utils.GetEmptyStruct(req.GetEntityType())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get empty struct for result")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	ctxWithReturnNew := driver.WithReturnNew(ctx, updatedEntity)
	meta, err := collection.UpdateDocument(ctxWithReturnNew, req.GetKey(), entityObj)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to update entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if err = utils.SetEntityMeta(req.GetEntityType(), updatedEntity, meta); err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to set entity meta")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	entity, err := utils.WrapEntity(req.GetEntityType(), updatedEntity)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to wrap entity")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.UpdateEntityResponse{
		Entity: entity,
	}, nil
}

func (s *EntityService) DeleteEntity(ctx context.Context, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to delete %s with ID: %s", userId, userRoles, req.GetEntityType(), req.GetKey())

	if !slices.Contains(userRoles, "admin") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	collection, err := s.getCollection(ctx, req.GetEntityType())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	_, err = collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to remove entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.DeleteEntityResponse{}, nil
}

func (s *EntityService) getCollection(ctx context.Context, name string) (driver.Collection, error) {
	s.mu.RLock()
	if col, ok := s.Collections[name]; ok {
		s.mu.RUnlock()
		return col, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check
	if col, ok := s.Collections[name]; ok {
		return col, nil
	}

	collection, err := s.DBClient.GetCreateCollection(ctx, name, driver.CreateVertexCollectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get or create %s collection: %v", name, err)
	}
	logrus.Infof("✅ Initialized collection %s", collection.Name())

	if name == "event" {
		collection.EnsurePersistentIndex(ctx, []string{"happened_at"}, &driver.EnsurePersistentIndexOptions{
			InBackground: true,
		})
	}

	s.Collections[name] = collection
	return collection, nil
}
