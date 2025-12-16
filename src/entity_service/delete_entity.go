package entityservice

import (
	"context"

	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omndapi/src/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EntityService) DeleteEntity(ctx context.Context, req *dapi.DeleteEntityRequest) (*dapi.DeleteEntityResponse, error) {
	// =====================================================
	// Get Common Data
	// =====================================================
	userId, userRoles, err := s.Pipeline.GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	logger := utils.GetLogger(ctx)
	logger.Infof("[%s, %v] requests to delete entity", userId, userRoles)

	col, err := s.Pipeline.GetCollection(req.GetEntityType())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity type: %v", err)
	}

	existingStruct, err := s.Pipeline.CreateEntityStruct(req.GetEntityType())
	if err != nil {
		return nil, err
	}

	// =====================================================
	// Check permission
	// =====================================================
	_, err = s.Pipeline.ReadDocument(ctx, col, req.GetKey(), existingStruct)
	if err != nil {
		return nil, err
	}

	if err := s.Pipeline.CheckDeletePermission(existingStruct, userId); err != nil {
		return nil, err
	}

	// =====================================================
	// Delete document
	// =====================================================
	if err := s.Pipeline.DeleteDocument(ctx, col, req.GetKey()); err != nil {
		return nil, err
	}

	return &dapi.DeleteEntityResponse{}, nil
}
