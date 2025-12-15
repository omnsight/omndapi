package handlers

import (
	"context"
	"fmt"
	"slices"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrganizationHandler struct {
	col driver.Collection
}

func NewOrganizationHandler(client *utils.ArangoDBClient) (*OrganizationHandler, error) {
	ctx := context.Background()
	col, err := client.GetCreateCollection(ctx, "organization", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get or create organization collection: %v", err)
	}
	logrus.Infof("✅ Initialized collection %s", col.Name())

	return &OrganizationHandler{col: col}, nil
}

func (h *OrganizationHandler) GetOrganization(ctx context.Context, req *dapi.GetEntityRequest) (*dapi.GetEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to get organization with ID: %s", userId, userRoles, req.GetKey())

	organization := &model.Organization{}
	meta, err := h.col.ReadDocument(ctx, req.GetKey(), organization)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "organization",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	organization.Id = meta.ID.String()
	organization.Key = meta.Key
	organization.Rev = meta.Rev

	if !utils.CheckReadPermission(organization, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to access this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	return &dapi.GetEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Organization{Organization: organization}},
	}, nil
}

func (h *OrganizationHandler) CreateOrganization(ctx context.Context, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to create organization", userId, userRoles)

	if !slices.Contains(userRoles, "admin") && !slices.Contains(userRoles, "pro") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	organization := req.GetEntity().GetOrganization()
	if organization == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not a organization")
	}

	// Prepare entity
	organization.Owner = userId
	organization.Read = nil
	organization.Write = nil

	createdOrganization := &model.Organization{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, createdOrganization)
	meta, err := h.col.CreateDocument(ctxWithReturnNew, organization)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "organization",
			"error":       err,
		}).Error("failed to create entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	createdOrganization.Id = meta.ID.String()
	createdOrganization.Key = meta.Key
	createdOrganization.Rev = meta.Rev

	return &dapi.CreateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Organization{Organization: createdOrganization}},
	}, nil
}

func (h *OrganizationHandler) UpdateOrganization(ctx context.Context, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update organization with ID: %s", userId, userRoles, req.GetKey())

	organization := req.GetEntity().GetOrganization()
	if organization == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not a organization")
	}

	existingOrganization := &model.Organization{}
	_, err = h.col.ReadDocument(ctx, req.GetKey(), existingOrganization)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "organization",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckWritePermission(existingOrganization, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to update this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	// Prepare update - copy permissions
	organization.Owner = existingOrganization.Owner
	organization.Read = existingOrganization.Read
	organization.Write = existingOrganization.Write

	updatedOrganization := &model.Organization{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, updatedOrganization)
	meta, err := h.col.UpdateDocument(ctxWithReturnNew, req.GetKey(), organization)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "organization",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to update entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	updatedOrganization.Id = meta.ID.String()
	updatedOrganization.Key = meta.Key
	updatedOrganization.Rev = meta.Rev

	return &dapi.UpdateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Organization{Organization: updatedOrganization}},
	}, nil
}

func (h *OrganizationHandler) DeleteOrganization(ctx context.Context, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	userId, _, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s] requests to delete organization with ID: %s", userId, req.GetKey())

	existingOrganization := &model.Organization{}
	_, err = h.col.ReadDocument(ctx, req.GetKey(), existingOrganization)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "organization",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckDeletePermission(existingOrganization, userId) {
		logger.WithFields(logrus.Fields{
			"key":  req.GetKey(),
			"user": userId,
		}).Info("User does not have permission to delete this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: only owner can delete entity")
	}

	_, err = h.col.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "organization",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to remove entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.DeleteEntityResponse{}, nil
}
