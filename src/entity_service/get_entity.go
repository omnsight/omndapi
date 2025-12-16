package entityservice

import (
	"context"

	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EntityService) GetEntity(ctx context.Context, req *dapi.GetEntityRequest) (*dapi.GetEntityResponse, error) {
	// =====================================================
	// Get Common Data
	// =====================================================
	userId, userRoles, err := s.Pipeline.GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to get entity", userId, userRoles)

	col, err := s.Pipeline.GetCollection(req.GetEntityType())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity type: %v", err)
	}

	// =====================================================
	// Process and clean up input data
	// =====================================================
	targetStruct, err := s.Pipeline.CreateEntityStruct(req.GetEntityType())
	if err != nil {
		return nil, err
	}

	meta, err := s.Pipeline.ReadDocument(ctx, col, req.GetKey(), targetStruct)
	if err != nil {
		return nil, err
	}

	// =====================================================
	// Check permission
	// =====================================================
	if err := s.Pipeline.CheckReadPermission(targetStruct, userId, userRoles); err != nil {
		return nil, err
	}

	// =====================================================
	// Wrap response
	// =====================================================
	s.Pipeline.SetEntityMeta(targetStruct, meta.ID.String(), meta.Key, meta.Rev)
	responseEntity, err := s.Pipeline.WrapEntityResponse(targetStruct)
	if err != nil {
		return nil, err
	}

	return &dapi.GetEntityResponse{Entity: responseEntity}, nil
}
