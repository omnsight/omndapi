package pipeline

import (
	"context"
	"slices"

	"github.com/omnsight/omndapi/src/utils"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetAuthInfo extracts user ID and roles from the context.
func (w *Worker) GetAuthInfo(ctx context.Context) (string, []string, error) {
	userId, userRoles, err := utils.GetUser(ctx)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to retrieve user info from context")
		return "", nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}
	return userId, userRoles, nil
}

// CheckCreatePermission checks if the user has permission to create an entity.
func (w *Worker) CheckCreatePermission(userRoles []string) error {
	if slices.Contains(userRoles, "admin") || slices.Contains(userRoles, "pro") {
		return nil
	}
	return status.Errorf(codes.PermissionDenied, "Access denied: only admin or pro users can create resources")
}

// CheckReadPermission checks if the user has permission to read the entity.
func (w *Worker) CheckReadPermission(entity ConcereteEntityCommon, userId string, userRoles []string) error {
	if entity.GetOwner() == userId {
		return nil
	}

	readList := entity.GetRead()
	if slices.Contains(readList, userId) {
		return nil
	}
	if len(lo.Intersect(readList, userRoles)) > 0 {
		return nil
	}

	return status.Errorf(codes.PermissionDenied, "Access denied")
}

func (w *Worker) CheckWritePermission(entity ConcereteEntityCommon, userId string, userRoles []string) error {
	if entity.GetOwner() == userId {
		return nil
	}

	if entity.GetOwner() == userId {
		return nil
	}

	writeList := entity.GetWrite()
	if slices.Contains(writeList, userId) {
		return nil
	}
	if len(lo.Intersect(writeList, userRoles)) > 0 {
		return nil
	}

	return status.Errorf(codes.PermissionDenied, "Access denied: user does not have required permissions")
}

func (w *Worker) CheckDeletePermission(entity ConcereteEntityCommon, userId string) error {
	if entity.GetOwner() == userId {
		return nil
	}

	return status.Errorf(codes.PermissionDenied, "Access denied: only owner can delete entity")
}
