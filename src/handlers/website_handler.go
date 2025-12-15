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

func GetWebsite(ctx context.Context, col driver.Collection, req *dapi.GetEntityRequest) (*dapi.GetEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to get website with ID: %s", userId, userRoles, req.GetKey())

	website := &model.Website{}
	meta, err := col.ReadDocument(ctx, req.GetKey(), website)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "website",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	website.Id = meta.ID.String()
	website.Key = meta.Key
	website.Rev = meta.Rev

	if !utils.CheckReadPermission(website, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to access this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	return &dapi.GetEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Website{Website: website}},
	}, nil
}

func CreateWebsite(ctx context.Context, col driver.Collection, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to create website", userId, userRoles)

	if !slices.Contains(userRoles, "admin") && !slices.Contains(userRoles, "pro") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	website := req.GetEntity().GetWebsite()
	if website == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not a website")
	}

	// Prepare entity
	website.Owner = userId
	website.Read = nil
	website.Write = nil

	createdWebsite := &model.Website{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, createdWebsite)
	meta, err := col.CreateDocument(ctxWithReturnNew, website)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "website",
			"error":       err,
		}).Error("failed to create entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	createdWebsite.Id = meta.ID.String()
	createdWebsite.Key = meta.Key
	createdWebsite.Rev = meta.Rev

	return &dapi.CreateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Website{Website: createdWebsite}},
	}, nil
}

func UpdateWebsite(ctx context.Context, col driver.Collection, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update website with ID: %s", userId, userRoles, req.GetKey())

	website := req.GetEntity().GetWebsite()
	if website == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not a website")
	}

	existingWebsite := &model.Website{}
	_, err = col.ReadDocument(ctx, req.GetKey(), existingWebsite)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "website",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckWritePermission(existingWebsite, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to update this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	// Prepare update - copy permissions
	website.Owner = existingWebsite.Owner
	website.Read = existingWebsite.Read
	website.Write = existingWebsite.Write

	updatedWebsite := &model.Website{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, updatedWebsite)
	meta, err := col.UpdateDocument(ctxWithReturnNew, req.GetKey(), website)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "website",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to update entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	updatedWebsite.Id = meta.ID.String()
	updatedWebsite.Key = meta.Key
	updatedWebsite.Rev = meta.Rev

	return &dapi.UpdateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Website{Website: updatedWebsite}},
	}, nil
}

func DeleteWebsite(ctx context.Context, col driver.Collection, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	userId, _, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s] requests to delete website with ID: %s", userId, req.GetKey())

	existingWebsite := &model.Website{}
	_, err = col.ReadDocument(ctx, req.GetKey(), existingWebsite)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "website",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckDeletePermission(existingWebsite, userId) {
		logger.WithFields(logrus.Fields{
			"key":  req.GetKey(),
			"user": userId,
		}).Info("User does not have permission to delete this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: only owner can delete entity")
	}

	_, err = col.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "website",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to remove entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.DeleteEntityResponse{}, nil
}
