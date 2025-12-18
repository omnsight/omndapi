package pipeline

import (
	"os"
	"sync"

	"github.com/arangodb/go-driver"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ConcereteEntityCommon interface {
	GetOwner() string
	GetRead() []string
	GetWrite() []string
}

type Worker struct {
	collections    map[string]driver.Collection
	openaiClient   *openai.Client
	embeddingModel openai.EmbeddingModel
	mu             sync.RWMutex
}

func NewWorker() *Worker {
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseUrl := os.Getenv("OPENAI_BASE_URL")
	modelName := os.Getenv("EMBEDDING_MODEL")

	var client *openai.Client
	if apiKey != "" {
		config := openai.DefaultConfig(apiKey)
		if baseUrl != "" {
			config.BaseURL = baseUrl
		}
		client = openai.NewClientWithConfig(config)
	} else {
		logrus.Warn("OPENAI_API_KEY not set, embeddings will not be generated properly")
	}

	embeddingModel := openai.AdaEmbeddingV2
	if modelName != "" {
		embeddingModel = openai.EmbeddingModel(modelName)
	}

	return &Worker{
		collections:    make(map[string]driver.Collection),
		openaiClient:   client,
		embeddingModel: embeddingModel,
	}
}

func (w *Worker) RegisterCollection(entityType string, col driver.Collection) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.collections[entityType] = col
}
