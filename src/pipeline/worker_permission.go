package pipeline

import (
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (w *Worker) SetPermissions(entity ConcereteEntityCommon, userId string, isCreate bool) error {
	var owner string
	var read, write []string

	if isCreate {
		owner = userId
		read = entity.GetRead()
		write = entity.GetWrite()
	} else {
		// Update case
		// In update, owner is cleared to avoid overwriting existing owner
		owner = ""

		// Check if input owner matches user ID
		currentOwner := entity.GetOwner()
		if currentOwner != userId {
			// If not matching (or empty), clear read/write permissions
			read = []string{}
			write = []string{}
		} else {
			// If matching, keep existing permissions from input
			read = entity.GetRead()
			write = entity.GetWrite()
		}
	}

	// Assign back to entity using type switch
	switch v := entity.(type) {
	case *model.Event:
		v.Owner = owner
		v.Read = read
		v.Write = write
	case *model.Source:
		v.Owner = owner
		v.Read = read
		v.Write = write
	case *model.Website:
		v.Owner = owner
		v.Read = read
		v.Write = write
	case *model.Person:
		v.Owner = owner
		v.Read = read
		v.Write = write
	case *model.Organization:
		v.Owner = owner
		v.Read = read
		v.Write = write
	case *model.Relation:
		v.Owner = owner
		v.Read = read
		v.Write = write
	default:
		return status.Errorf(codes.InvalidArgument, "unknown entity type")
	}

	return nil
}
