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

func GetSource(ctx context.Context, col driver.Collection, req *dapi.GetEntityRequest) (*dapi.GetEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to get source with ID: %s", userId, userRoles, req.GetKey())

	source := &model.Source{}
	meta, err := col.ReadDocument(ctx, req.GetKey(), source)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "source",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	source.Id = meta.ID.String()
	source.Key = meta.Key
	source.Rev = meta.Rev

	if !utils.CheckReadPermission(source, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to access this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	return &dapi.GetEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Source{Source: source}},
	}, nil
}

func CreateSource(ctx context.Context, col driver.Collection, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to create source", userId, userRoles)

	if !slices.Contains(userRoles, "admin") && !slices.Contains(userRoles, "pro") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	source := req.GetEntity().GetSource()
	if source == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not a source")
	}

	// Prepare entity
	source.Owner = userId
	source.Read = nil
	source.Write = nil

	createdSource := &model.Source{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, createdSource)
	meta, err := col.CreateDocument(ctxWithReturnNew, source)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "source",
			"error":       err,
		}).Error("failed to create entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	createdSource.Id = meta.ID.String()
	createdSource.Key = meta.Key
	createdSource.Rev = meta.Rev

	return &dapi.CreateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Source{Source: createdSource}},
	}, nil
}

func UpdateSource(ctx context.Context, col driver.Collection, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update source with ID: %s", userId, userRoles, req.GetKey())

	source := req.GetEntity().GetSource()
	if source == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not a source")
	}

	existingSource := &model.Source{}
	_, err = col.ReadDocument(ctx, req.GetKey(), existingSource)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "source",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckWritePermission(existingSource, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to update this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	// Prepare update - copy permissions
	source.Owner = existingSource.Owner
	source.Read = existingSource.Read
	source.Write = existingSource.Write

	updatedSource := &model.Source{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, updatedSource)
	meta, err := col.UpdateDocument(ctxWithReturnNew, req.GetKey(), source)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "source",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to update entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	updatedSource.Id = meta.ID.String()
	updatedSource.Key = meta.Key
	updatedSource.Rev = meta.Rev

	return &dapi.UpdateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Source{Source: updatedSource}},
	}, nil
}

func DeleteSource(ctx context.Context, col driver.Collection, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	userId, _, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s] requests to delete source with ID: %s", userId, req.GetKey())

	existingSource := &model.Source{}
	_, err = col.ReadDocument(ctx, req.GetKey(), existingSource)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "source",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckDeletePermission(existingSource, userId) {
		logger.WithFields(logrus.Fields{
			"key":  req.GetKey(),
			"user": userId,
		}).Info("User does not have permission to delete this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: only owner can delete entity")
	}

	_, err = col.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "source",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to remove entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.DeleteEntityResponse{}, nil
}
