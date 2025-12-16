package pipeline

import (
	"reflect"

	"github.com/omnsight/omniscent-library/gen/model/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (w *Worker) WrapEntityResponse(entity interface{}) (*model.Entity, error) {
	switch v := entity.(type) {
	case *model.Event:
		return &model.Entity{Entity: &model.Entity_Event{Event: v}}, nil
	case *model.Source:
		return &model.Entity{Entity: &model.Entity_Source{Source: v}}, nil
	case *model.Website:
		return &model.Entity{Entity: &model.Entity_Website{Website: v}}, nil
	case *model.Person:
		return &model.Entity{Entity: &model.Entity_Person{Person: v}}, nil
	case *model.Organization:
		return &model.Entity{Entity: &model.Entity_Organization{Organization: v}}, nil
	default:
		return nil, status.Errorf(codes.Internal, "unknown entity type for response wrapping")
	}
}

func (w *Worker) SetEntityMeta(entity interface{}, id, key, rev string) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	if f := v.FieldByName("Id"); f.IsValid() && f.CanSet() {
		f.SetString(id)
	}
	if f := v.FieldByName("Key"); f.IsValid() && f.CanSet() {
		f.SetString(key)
	}
	if f := v.FieldByName("Rev"); f.IsValid() && f.CanSet() {
		f.SetString(rev)
	}
}
