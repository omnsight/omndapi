package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/handlers"
	"github.com/omnsight/omndapi/src/utils"
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
				FILTER (
					v.owner == @userId OR 
					@userId IN v.read OR 
					LENGTH(INTERSECTION(@userRoles, v.read)) > 0
				)
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
		"userId":      userId,
		"userRoles":   userRoles,
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

	var result handlers.QueryResult
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

	pbEntities, pbRelations := handlers.ProcessEntities(ctx, result.Entities, result.Relations, userId, userRoles)

	return &dapi.ListEntitiesFromEventResponse{
		Entities:  pbEntities,
		Relations: pbRelations,
	}, nil
}

func (s *EntityService) GetEntity(ctx context.Context, req *dapi.GetEntityRequest) (*dapi.GetEntityResponse, error) {
	collection, err := s.getCollection(ctx, req.GetEntityType())
	if err != nil {
		utils.GetLogger(ctx).WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	switch req.GetEntityType() {
	case "event":
		return handlers.GetEvent(ctx, collection, req)
	case "source":
		return handlers.GetSource(ctx, collection, req)
	case "website":
		return handlers.GetWebsite(ctx, collection, req)
	case "person":
		return handlers.GetPerson(ctx, collection, req)
	case "organization":
		return handlers.GetOrganization(ctx, collection, req)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
}

func (s *EntityService) CreateEntity(ctx context.Context, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	collection, err := s.getCollection(ctx, req.GetEntityType())
	if err != nil {
		utils.GetLogger(ctx).WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	switch req.GetEntityType() {
	case "event":
		return handlers.CreateEvent(ctx, collection, req)
	case "source":
		return handlers.CreateSource(ctx, collection, req)
	case "website":
		return handlers.CreateWebsite(ctx, collection, req)
	case "person":
		return handlers.CreatePerson(ctx, collection, req)
	case "organization":
		return handlers.CreateOrganization(ctx, collection, req)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
}

func (s *EntityService) UpdateEntity(ctx context.Context, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	collection, err := s.getCollection(ctx, req.GetEntityType())
	if err != nil {
		utils.GetLogger(ctx).WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	switch req.GetEntityType() {
	case "event":
		return handlers.UpdateEvent(ctx, collection, req)
	case "source":
		return handlers.UpdateSource(ctx, collection, req)
	case "website":
		return handlers.UpdateWebsite(ctx, collection, req)
	case "person":
		return handlers.UpdatePerson(ctx, collection, req)
	case "organization":
		return handlers.UpdateOrganization(ctx, collection, req)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
}

func (s *EntityService) DeleteEntity(ctx context.Context, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	collection, err := s.getCollection(ctx, req.GetEntityType())
	if err != nil {
		utils.GetLogger(ctx).WithFields(logrus.Fields{
			"entity_type": req.GetEntityType(),
			"error":       err,
		}).Error("failed to get collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	switch req.GetEntityType() {
	case "event":
		return handlers.DeleteEvent(ctx, collection, req)
	case "source":
		return handlers.DeleteSource(ctx, collection, req)
	case "website":
		return handlers.DeleteWebsite(ctx, collection, req)
	case "person":
		return handlers.DeletePerson(ctx, collection, req)
	case "organization":
		return handlers.DeleteOrganization(ctx, collection, req)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
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
