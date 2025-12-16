package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateDocument inserts the document into the collection and unmarshals the result into resultStruct.
func (w *Worker) CreateDocument(ctx context.Context, col driver.Collection, data map[string]interface{}, resultStruct interface{}) (driver.DocumentMeta, error) {
	var resultMap map[string]interface{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, &resultMap)

	meta, err := col.CreateDocument(ctxWithReturnNew, data)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"collection": col.Name(),
			"error":      err,
		}).Error("Failed to create document")
		return driver.DocumentMeta{}, status.Errorf(codes.Internal, "Internal service error")
	}

	if err := w.mapToStruct(resultMap, resultStruct); err != nil {
		return meta, err
	}
	return meta, nil
}

// ReadDocument reads a document by key into resultStruct.
func (w *Worker) ReadDocument(ctx context.Context, col driver.Collection, key string, resultStruct interface{}) (driver.DocumentMeta, error) {
	var resultMap map[string]interface{}
	meta, err := col.ReadDocument(ctx, key, &resultMap)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return driver.DocumentMeta{}, status.Errorf(codes.NotFound, "entity not found")
		}
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"collection": col.Name(),
			"key":        key,
			"error":      err,
		}).Error("Failed to read document")
		return driver.DocumentMeta{}, status.Errorf(codes.Internal, "Internal service error")
	}

	if err := w.mapToStruct(resultMap, resultStruct); err != nil {
		return meta, err
	}
	return meta, nil
}

// UpdateDocument updates a document and unmarshals the result into resultStruct.
func (w *Worker) UpdateDocument(ctx context.Context, col driver.Collection, key string, data map[string]interface{}, resultStruct interface{}) (driver.DocumentMeta, error) {
	var resultMap map[string]interface{}
	ctxWithReturnNew := driver.WithReturnNew(ctx, &resultMap)

	meta, err := col.UpdateDocument(ctxWithReturnNew, key, data)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"collection": col.Name(),
			"key":        key,
			"error":      err,
		}).Error("Failed to update document")
		return driver.DocumentMeta{}, status.Errorf(codes.Internal, "Internal service error")
	}

	if err := w.mapToStruct(resultMap, resultStruct); err != nil {
		return meta, err
	}
	return meta, nil
}

// DeleteDocument removes a document by key.
func (w *Worker) DeleteDocument(ctx context.Context, col driver.Collection, key string) error {
	_, err := col.RemoveDocument(ctx, key)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"collection": col.Name(),
			"key":        key,
			"error":      err,
		}).Error("Failed to delete document")
		return status.Errorf(codes.Internal, "Internal service error")
	}
	return nil
}

func (w *Worker) mapToStruct(m map[string]interface{}, v interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal into struct: %w", err)
	}
	return nil
}
