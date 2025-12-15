package main

import (
	"context"

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

	DBClient *utils.ArangoDBClient

	EventHandler        *handlers.EventHandler
	SourceHandler       *handlers.SourceHandler
	WebsiteHandler      *handlers.WebsiteHandler
	PersonHandler       *handlers.PersonHandler
	OrganizationHandler *handlers.OrganizationHandler
}

func NewEntityService(client *utils.ArangoDBClient) (*EntityService, error) {
	service := &EntityService{
		DBClient: client,
	}

	var err error

	// Initialize handlers
	service.EventHandler, err = handlers.NewEventHandler(client)
	if err != nil {
		return nil, err
	}

	service.SourceHandler, err = handlers.NewSourceHandler(client)
	if err != nil {
		return nil, err
	}

	service.WebsiteHandler, err = handlers.NewWebsiteHandler(client)
	if err != nil {
		return nil, err
	}

	service.PersonHandler, err = handlers.NewPersonHandler(client)
	if err != nil {
		return nil, err
	}

	service.OrganizationHandler, err = handlers.NewOrganizationHandler(client)
	if err != nil {
		return nil, err
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
	switch req.GetEntityType() {
	case "event":
		return s.EventHandler.GetEvent(ctx, req)
	case "source":
		return s.SourceHandler.GetSource(ctx, req)
	case "website":
		return s.WebsiteHandler.GetWebsite(ctx, req)
	case "person":
		return s.PersonHandler.GetPerson(ctx, req)
	case "organization":
		return s.OrganizationHandler.GetOrganization(ctx, req)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
}

func (s *EntityService) CreateEntity(ctx context.Context, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	switch req.GetEntityType() {
	case "event":
		return s.EventHandler.CreateEvent(ctx, req)
	case "source":
		return s.SourceHandler.CreateSource(ctx, req)
	case "website":
		return s.WebsiteHandler.CreateWebsite(ctx, req)
	case "person":
		return s.PersonHandler.CreatePerson(ctx, req)
	case "organization":
		return s.OrganizationHandler.CreateOrganization(ctx, req)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
}

func (s *EntityService) UpdateEntity(ctx context.Context, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	switch req.GetEntityType() {
	case "event":
		return s.EventHandler.UpdateEvent(ctx, req)
	case "source":
		return s.SourceHandler.UpdateSource(ctx, req)
	case "website":
		return s.WebsiteHandler.UpdateWebsite(ctx, req)
	case "person":
		return s.PersonHandler.UpdatePerson(ctx, req)
	case "organization":
		return s.OrganizationHandler.UpdateOrganization(ctx, req)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
}

func (s *EntityService) DeleteEntity(ctx context.Context, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	switch req.GetEntityType() {
	case "event":
		return s.EventHandler.DeleteEvent(ctx, req)
	case "source":
		return s.SourceHandler.DeleteSource(ctx, req)
	case "website":
		return s.WebsiteHandler.DeleteWebsite(ctx, req)
	case "person":
		return s.PersonHandler.DeletePerson(ctx, req)
	case "organization":
		return s.OrganizationHandler.DeleteOrganization(ctx, req)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
}
