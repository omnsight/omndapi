package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (w *Worker) SetAdditionalFields(entity interface{}) (map[string]interface{}, error) {
	var textParts []string

	switch v := entity.(type) {
	case *model.Event:
		textParts = append(textParts, v.GetDescription())
		textParts = append(textParts, v.GetTags()...)
	case *model.Source:
		textParts = append(textParts, v.GetDescription())
	case *model.Website:
		textParts = append(textParts, v.GetTitle())
		textParts = append(textParts, v.GetUrl())
	case *model.Person:
		textParts = append(textParts, v.GetName())
	case *model.Organization:
		textParts = append(textParts, v.GetName())
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown entity type")
	}

	combinedText := strings.Join(textParts, " ")
	_ = combinedText // Placeholder for actual embedding generation using combinedText

	const vectorSize = 1536
	vector := make([]float32, vectorSize)

	for i := 0; i < 10 && i < vectorSize; i++ {
		vector[i] = 0.1 * float32(i+1)
	}

	logrus.Debugf("Generated mock vector based on text length %d", len(combinedText))

	// Convert entity to map
	data, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	var entityMap map[string]interface{}
	if err := json.Unmarshal(data, &entityMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into map: %w", err)
	}

	entityMap["vector"] = vector
	return entityMap, nil
}
