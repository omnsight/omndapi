package handlers

import (
	"context"
	"slices"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetEvent(ctx context.Context, col driver.Collection, req *dapi.GetEntityRequest) (*dapi.GetEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to get event with ID: %s", userId, userRoles, req.GetKey())

	event := &model.Event{}
	meta, err := col.ReadDocument(ctx, req.GetKey(), event)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"entity_type": "event",
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

	event.Id = meta.ID.String()
	event.Key = meta.Key
	event.Rev = meta.Rev

	if !utils.CheckReadPermission(event, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to access this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	return &dapi.GetEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Event{Event: event}},
	}, nil
}

func CreateEvent(ctx context.Context, col driver.Collection, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to create event", userId, userRoles)

	if !slices.Contains(userRoles, "admin") && !slices.Contains(userRoles, "pro") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	event := req.GetEntity().GetEvent()
	if event == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not an event")
	}

	// Prepare entity
	event.Owner = userId
	event.Read = nil
	event.Write = nil

	createdEvent := &model.Event{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, createdEvent)
	meta, err := col.CreateDocument(ctxWithReturnNew, event)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "event",
			"error":       err,
		}).Error("failed to create entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	createdEvent.Id = meta.ID.String()
	createdEvent.Key = meta.Key
	createdEvent.Rev = meta.Rev

	return &dapi.CreateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Event{Event: createdEvent}},
	}, nil
}

func UpdateEvent(ctx context.Context, col driver.Collection, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update event with ID: %s", userId, userRoles, req.GetKey())

	event := req.GetEntity().GetEvent()
	if event == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not an event")
	}

	existingEvent := &model.Event{}
	_, err = col.ReadDocument(ctx, req.GetKey(), existingEvent)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "event",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckWritePermission(existingEvent, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to update this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	// Prepare update - copy permissions
	event.Owner = existingEvent.Owner
	event.Read = existingEvent.Read
	event.Write = existingEvent.Write

	updatedEvent := &model.Event{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, updatedEvent)
	meta, err := col.UpdateDocument(ctxWithReturnNew, req.GetKey(), event)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "event",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to update entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	updatedEvent.Id = meta.ID.String()
	updatedEvent.Key = meta.Key
	updatedEvent.Rev = meta.Rev

	return &dapi.UpdateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Event{Event: updatedEvent}},
	}, nil
}

func DeleteEvent(ctx context.Context, col driver.Collection, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	userId, _, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s] requests to delete event with ID: %s", userId, req.GetKey())

	existingEvent := &model.Event{}
	_, err = col.ReadDocument(ctx, req.GetKey(), existingEvent)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "event",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckDeletePermission(existingEvent, userId) {
		logger.WithFields(logrus.Fields{
			"key":  req.GetKey(),
			"user": userId,
		}).Info("User does not have permission to delete this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: only owner can delete entity")
	}

	_, err = col.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "event",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to remove entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.DeleteEntityResponse{}, nil
}
