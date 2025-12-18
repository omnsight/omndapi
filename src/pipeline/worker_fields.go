package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/omnsight/omniscent-library/gen/model/v1"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (w *Worker) SetAdditionalFields(ctx context.Context, entity interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	var entityMap map[string]interface{}
	if err := json.Unmarshal(data, &entityMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into map: %w", err)
	}

	embeddings, err := w.GetEmbedding(ctx, entity)
	if err != nil {
		return nil, err
	}
	entityMap["embedding"] = embeddings
	return entityMap, nil
}

func (w *Worker) GetEmbedding(ctx context.Context, entity interface{}) ([]float32, error) {
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

	if w.openaiClient == nil {
		logrus.Debug("OpenAI client not initialized, returning mock embedding")
		const vectorSize = 1536
		embeddings := make([]float32, vectorSize)

		for i := 0; i < 10 && i < vectorSize; i++ {
			embeddings[i] = 0.1 * float32(i+1)
		}

		logrus.Debugf("Generated mock vector based on text length %d", len(combinedText))
		return embeddings, nil
	}

	resp, err := w.openaiClient.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: []string{combinedText},
			Model: w.embeddingModel,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return resp.Data[0].Embedding, nil
}
