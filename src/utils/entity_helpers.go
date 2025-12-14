package utils

import (
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omniscent-library/gen/model/v1"
)

// Helper functions for entity handling
func GetEmptyStruct(entityType string) (interface{}, error) {
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
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}
}

func UnwrapEntity(entity *model.Entity) (string, interface{}, error) {
	if e := entity.GetEvent(); e != nil {
		return "event", e, nil
	}
	if e := entity.GetSource(); e != nil {
		return "source", e, nil
	}
	if e := entity.GetWebsite(); e != nil {
		return "website", e, nil
	}
	if e := entity.GetPerson(); e != nil {
		return "person", e, nil
	}
	if e := entity.GetOrganization(); e != nil {
		return "organization", e, nil
	}
	return "", nil, fmt.Errorf("empty or unknown entity content")
}

func WrapEntity(entityType string, obj interface{}) (*model.Entity, error) {
	switch entityType {
	case "event":
		if v, ok := obj.(*model.Event); ok {
			return &model.Entity{Entity: &model.Entity_Event{Event: v}}, nil
		}
	case "source":
		if v, ok := obj.(*model.Source); ok {
			return &model.Entity{Entity: &model.Entity_Source{Source: v}}, nil
		}
	case "website":
		if v, ok := obj.(*model.Website); ok {
			return &model.Entity{Entity: &model.Entity_Website{Website: v}}, nil
		}
	case "person":
		if v, ok := obj.(*model.Person); ok {
			return &model.Entity{Entity: &model.Entity_Person{Person: v}}, nil
		}
	case "organization":
		if v, ok := obj.(*model.Organization); ok {
			return &model.Entity{Entity: &model.Entity_Organization{Organization: v}}, nil
		}
	}
	return nil, fmt.Errorf("failed to wrap entity: type mismatch for %s", entityType)
}

func SetEntityMeta(entityType string, obj interface{}, meta driver.DocumentMeta) error {
	switch entityType {
	case "event":
		if v, ok := obj.(*model.Event); ok {
			v.Id = meta.ID.String()
			v.Key = meta.Key
			v.Rev = meta.Rev
			return nil
		}
	case "source":
		if v, ok := obj.(*model.Source); ok {
			v.Id = meta.ID.String()
			v.Key = meta.Key
			v.Rev = meta.Rev
			return nil
		}
	case "website":
		if v, ok := obj.(*model.Website); ok {
			v.Id = meta.ID.String()
			v.Key = meta.Key
			v.Rev = meta.Rev
			return nil
		}
	case "person":
		if v, ok := obj.(*model.Person); ok {
			v.Id = meta.ID.String()
			v.Key = meta.Key
			v.Rev = meta.Rev
			return nil
		}
	case "organization":
		if v, ok := obj.(*model.Organization); ok {
			v.Id = meta.ID.String()
			v.Key = meta.Key
			v.Rev = meta.Rev
			return nil
		}
	}
	return fmt.Errorf("failed to set entity meta: type mismatch for %s", entityType)
}

func GetEntityRoles(entityType string, obj interface{}) ([]string, error) {
	// Define an interface for the GetRoles method which is common to all entity types
	type roleGetter interface {
		GetRoles() []string
	}

	if v, ok := obj.(roleGetter); ok {
		return v.GetRoles(), nil
	}

	return nil, fmt.Errorf("failed to get entity roles: type mismatch for %s or method not available", entityType)
}
