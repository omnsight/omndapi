package entityservice

import (
	"context"

	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EntityService) CreateEntity(ctx context.Context, req *dapi.CreateEntityRequest) (*dapi.CreateEntityResponse, error) {
	// =====================================================
	// Get Common Data
	// =====================================================
	userId, userRoles, err := s.Pipeline.GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to create entity", userId, userRoles)

	if err := s.Pipeline.CheckCreatePermission(userRoles); err != nil {
		return nil, err
	}

	col, err := s.Pipeline.GetCollection(req.GetEntityType())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity type: %v", err)
	}

	inputEntity, err := s.Pipeline.ExtractInputEntity(req)
	if err != nil {
		return nil, err
	}

	if inputEntity == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity content missing")
	}

	// =====================================================
	// Process and clean up input data
	// =====================================================
	if err := s.Pipeline.SetPermissions(inputEntity, userId, true); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set permissions: %v", err)
	}

	dataMap, err := s.Pipeline.SetAdditionalFields(ctx, inputEntity)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set additional fields: %v", err)
	}

	createdStruct, err := s.Pipeline.CreateEntityStruct(req.GetEntityType())
	if err != nil {
		return nil, err
	}

	// =====================================================
	// Write into db
	// =====================================================
	meta, err := s.Pipeline.CreateDocument(ctx, col, dataMap, createdStruct)
	if err != nil {
		return nil, err
	}

	// =====================================================
	// Convert back to response entity
	// =====================================================
	s.Pipeline.SetEntityMeta(createdStruct, meta.ID.String(), meta.Key, meta.Rev)
	responseEntity, err := s.Pipeline.WrapEntityResponse(createdStruct)
	if err != nil {
		return nil, err
	}

	return &dapi.CreateEntityResponse{Entity: responseEntity}, nil
}
