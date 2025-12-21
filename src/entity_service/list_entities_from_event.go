package entityservice

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/pipeline"
	"github.com/omnsight/omndapi/src/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EntityService) ListEntitiesFromEvent(ctx context.Context, req *dapi.ListEntitiesFromEventRequest) (*dapi.ListEntitiesFromEventResponse, error) {
	userId, userRoles, err := s.Pipeline.GetAuthInfo(ctx)
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
				SORT e.happened_at DESC
				RETURN e
			)
		)

		LET filtered_events = (
			FOR e IN start_events
			FILTER (@countryCode == "" OR e.location.countryCode == @countryCode OR e.location.country_code == @countryCode)
			FILTER (@tag == "" 
				OR @tag IN e.tags 
				OR (IS_DOCUMENT(e.attributes) AND LENGTH(
					FOR lang IN ATTRIBUTES(e.attributes)
					FILTER IS_LIST(e.attributes[lang].Tags) AND @tag IN e.attributes[lang].Tags
					RETURN 1
				) > 0)
			)
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

	var result pipeline.QueryResult
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

	pbEntities, pbRelations := s.Pipeline.ProcessEntities(ctx, result.Entities, result.Relations)

	return &dapi.ListEntitiesFromEventResponse{
		Entities:  pbEntities,
		Relations: pbRelations,
	}, nil
}
