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

func GetPerson(ctx context.Context, col driver.Collection, req *dapi.GetEntityRequest) (*dapi.GetEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to get person with ID: %s", userId, userRoles, req.GetKey())

	person := &model.Person{}
	meta, err := col.ReadDocument(ctx, req.GetKey(), person)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "person",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	person.Id = meta.ID.String()
	person.Key = meta.Key
	person.Rev = meta.Rev

	if !utils.CheckReadPermission(person, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to access this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	return &dapi.GetEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Person{Person: person}},
	}, nil
}

func CreatePerson(ctx context.Context, col driver.Collection, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to create person", userId, userRoles)

	if !slices.Contains(userRoles, "admin") && !slices.Contains(userRoles, "pro") {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized! You don't have permission.")
	}

	person := req.GetEntity().GetPerson()
	if person == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not a person")
	}

	// Prepare entity
	person.Owner = userId
	person.Read = nil
	person.Write = nil

	createdPerson := &model.Person{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, createdPerson)
	meta, err := col.CreateDocument(ctxWithReturnNew, person)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "person",
			"error":       err,
		}).Error("failed to create entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	createdPerson.Id = meta.ID.String()
	createdPerson.Key = meta.Key
	createdPerson.Rev = meta.Rev

	return &dapi.CreateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Person{Person: createdPerson}},
	}, nil
}

func UpdatePerson(ctx context.Context, col driver.Collection, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update person with ID: %s", userId, userRoles, req.GetKey())

	person := req.GetEntity().GetPerson()
	if person == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing or not a person")
	}

	existingPerson := &model.Person{}
	_, err = col.ReadDocument(ctx, req.GetKey(), existingPerson)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "person",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckWritePermission(existingPerson, userId, userRoles) {
		logger.WithFields(logrus.Fields{
			"key":        req.GetKey(),
			"user":       userId,
			"user_roles": userRoles,
		}).Info("User does not have permission to update this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
	}

	// Prepare update - copy permissions
	person.Owner = existingPerson.Owner
	person.Read = existingPerson.Read
	person.Write = existingPerson.Write

	updatedPerson := &model.Person{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, updatedPerson)
	meta, err := col.UpdateDocument(ctxWithReturnNew, req.GetKey(), person)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "person",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to update entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	updatedPerson.Id = meta.ID.String()
	updatedPerson.Key = meta.Key
	updatedPerson.Rev = meta.Rev

	return &dapi.UpdateEntityResponse{
		Entity: &model.Entity{Entity: &model.Entity_Person{Person: updatedPerson}},
	}, nil
}

func DeletePerson(ctx context.Context, col driver.Collection, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	userId, _, err := utils.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s] requests to delete person with ID: %s", userId, req.GetKey())

	existingPerson := &model.Person{}
	_, err = col.ReadDocument(ctx, req.GetKey(), existingPerson)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, status.Errorf(codes.NotFound, "entity not found")
		}
		logger.WithFields(logrus.Fields{
			"entity_type": "person",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to read existing entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !utils.CheckDeletePermission(existingPerson, userId) {
		logger.WithFields(logrus.Fields{
			"key":  req.GetKey(),
			"user": userId,
		}).Info("User does not have permission to delete this entity.")
		return nil, status.Errorf(codes.PermissionDenied, "Access denied: only owner can delete entity")
	}

	_, err = col.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"entity_type": "person",
			"key":         req.GetKey(),
			"error":       err,
		}).Error("failed to remove entity document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &dapi.DeleteEntityResponse{}, nil
}
