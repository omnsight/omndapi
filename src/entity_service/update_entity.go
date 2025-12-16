package entityservice

import (
	"context"

	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EntityService) UpdateEntity(ctx context.Context, req *dapi.UpdateEntityRequest) (*dapi.UpdateEntityResponse, error) {
	// =====================================================
	// Get Common Data
	// =====================================================
	userId, userRoles, err := s.Pipeline.GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to update entity", userId, userRoles)

	col, err := s.Pipeline.GetCollection(req.GetEntityType())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity type: %v", err)
	}

	existingStruct, err := s.Pipeline.CreateEntityStruct(req.GetEntityType())
	if err != nil {
		return nil, err
	}
	inputEntity, err := s.Pipeline.ExtractInputEntity(req)
	if err != nil {
		return nil, err
	}

	if inputEntity == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing")
	}

	// =====================================================
	// Check Permission
	// =====================================================
	_, err = s.Pipeline.ReadDocument(ctx, col, req.GetKey(), existingStruct)
	if err != nil {
		return nil, err
	}

	if err := s.Pipeline.CheckWritePermission(existingStruct, userId, userRoles); err != nil {
		return nil, err
	}

	// =====================================================
	// Process and clean up input data
	// =====================================================
	if err := s.Pipeline.SetPermissions(inputEntity, userId, false); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set permissions: %v", err)
	}

	dataMap, err := s.Pipeline.SetAdditionalFields(inputEntity)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set additional fields: %v", err)
	}

	updatedStruct, err := s.Pipeline.CreateEntityStruct(req.GetEntityType())
	if err != nil {
		return nil, err
	}

	// =====================================================
	// Write into db
	// =====================================================
	meta, err := s.Pipeline.UpdateDocument(ctx, col, req.GetKey(), dataMap, updatedStruct)
	if err != nil {
		return nil, err
	}

	// =====================================================
	// Return response
	// =====================================================
	s.Pipeline.SetEntityMeta(updatedStruct, meta.ID.String(), meta.Key, meta.Rev)
	responseEntity, err := s.Pipeline.WrapEntityResponse(updatedStruct)
	if err != nil {
		return nil, err
	}

	return &dapi.UpdateEntityResponse{Entity: responseEntity}, nil
}
