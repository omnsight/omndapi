package pipeline

import (
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EntityRequest is a common interface for CreateEntityRequest and UpdateEntityRequest
type EntityRequest interface {
	GetEntityType() string
	GetEntity() *model.Entity
}

func (w *Worker) CreateEntityStruct(entityType string) (ConcereteEntityCommon, error) {
	switch entityType {
	case "event":
		return &model.Event{}, nil
	case "source":
		return &model.Source{}, nil
	case "website":
		return &model.Website{}, nil
	case "person":
		return &model.Person{}, nil
	case "organization":
		return &model.Organization{}, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", entityType)
	}
}

func (w *Worker) ExtractInputEntity(req EntityRequest) (ConcereteEntityCommon, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request is nil")
	}

	wrapper := req.GetEntity()
	if wrapper == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity wrapper is nil")
	}

	switch req.GetEntityType() {
	case "event":
		return wrapper.GetEvent(), nil
	case "source":
		return wrapper.GetSource(), nil
	case "website":
		return wrapper.GetWebsite(), nil
	case "person":
		return wrapper.GetPerson(), nil
	case "organization":
		return wrapper.GetOrganization(), nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type: %s", req.GetEntityType())
	}
}
