package handlers

import (
	"context"
	"encoding/json"

	"github.com/omnsight/omndapi/src/utils"
	"github.com/omnsight/omniscent-library/gen/model/v1"
)

type EntityResult struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type QueryResult struct {
	Entities  []EntityResult   `json:"entities"`
	Relations []model.Relation `json:"relations"`
}

func ProcessEntities(ctx context.Context, entities []EntityResult, relations []model.Relation, userId string, userRoles []string) ([]*model.Entity, []*model.Relation) {
	var pbEntities []*model.Entity
	allowedIds := make(map[string]struct{})
	logger := utils.GetLogger(ctx)

	for _, er := range entities {
		switch er.Type {
		case "event":
			var event model.Event
			if err := json.Unmarshal(er.Data, &event); err != nil {
				logger.Errorf("failed to unmarshal entity data for type %s: %v", er.Type, err)
				continue
			}
			if !utils.CheckReadPermission(&event, userId, userRoles) {
				continue
			}
			allowedIds[event.GetId()] = struct{}{}
			pbEntities = append(pbEntities, &model.Entity{Entity: &model.Entity_Event{Event: &event}})

		case "source":
			var source model.Source
			if err := json.Unmarshal(er.Data, &source); err != nil {
				logger.Errorf("failed to unmarshal entity data for type %s: %v", er.Type, err)
				continue
			}
			if !utils.CheckReadPermission(&source, userId, userRoles) {
				continue
			}
			allowedIds[source.GetId()] = struct{}{}
			pbEntities = append(pbEntities, &model.Entity{Entity: &model.Entity_Source{Source: &source}})

		case "website":
			var website model.Website
			if err := json.Unmarshal(er.Data, &website); err != nil {
				logger.Errorf("failed to unmarshal entity data for type %s: %v", er.Type, err)
				continue
			}
			if !utils.CheckReadPermission(&website, userId, userRoles) {
				continue
			}
			allowedIds[website.GetId()] = struct{}{}
			pbEntities = append(pbEntities, &model.Entity{Entity: &model.Entity_Website{Website: &website}})

		case "person":
			var person model.Person
			if err := json.Unmarshal(er.Data, &person); err != nil {
				logger.Errorf("failed to unmarshal entity data for type %s: %v", er.Type, err)
				continue
			}
			if !utils.CheckReadPermission(&person, userId, userRoles) {
				continue
			}
			allowedIds[person.GetId()] = struct{}{}
			pbEntities = append(pbEntities, &model.Entity{Entity: &model.Entity_Person{Person: &person}})

		case "organization":
			var organization model.Organization
			if err := json.Unmarshal(er.Data, &organization); err != nil {
				logger.Errorf("failed to unmarshal entity data for type %s: %v", er.Type, err)
				continue
			}
			if !utils.CheckReadPermission(&organization, userId, userRoles) {
				continue
			}
			allowedIds[organization.GetId()] = struct{}{}
			pbEntities = append(pbEntities, &model.Entity{Entity: &model.Entity_Organization{Organization: &organization}})

		default:
			logger.Warnf("unknown entity type %s", er.Type)
		}
	}

	var pbRelations []*model.Relation
	for i := range relations {
		r := &relations[i]
		_, fromOk := allowedIds[r.From]
		_, toOk := allowedIds[r.To]
		if fromOk && toOk {
			pbRelations = append(pbRelations, r)
		}
	}

	return pbEntities, pbRelations
}
