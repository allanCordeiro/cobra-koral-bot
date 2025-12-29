package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/allan/cobra-coral/internal/domain"
)

// GCSStateRepository implements domain.StateRepository using Google Cloud Storage
type GCSStateRepository struct {
	client     *storage.Client
	bucketName string
	objectName string
	logger     domain.Logger
	ctx        context.Context
}

// NewGCSStateRepository creates a new GCS-based state repository
func NewGCSStateRepository(ctx context.Context, bucketName, objectName string, logger domain.Logger) (*GCSStateRepository, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStateRepository{
		client:     client,
		bucketName: bucketName,
		objectName: objectName,
		logger:     logger,
		ctx:        ctx,
	}, nil
}

// Close closes the GCS client
func (r *GCSStateRepository) Close() error {
	return r.client.Close()
}

// GetLastExecution retrieves the last execution state from Cloud Storage
func (r *GCSStateRepository) GetLastExecution() (*domain.ExecutionState, error) {
	r.logger.Log(domain.INFO, fmt.Sprintf("Reading state from gs://%s/%s", r.bucketName, r.objectName), "GCSStateRepo")

	bucket := r.client.Bucket(r.bucketName)
	obj := bucket.Object(r.objectName)

	// Try to read the object
	reader, err := obj.NewReader(r.ctx)
	if err != nil {
		// If object doesn't exist, return zero time (first execution)
		if err == storage.ErrObjectNotExist {
			r.logger.Log(domain.INFO, "State file doesn't exist yet, returning zero time", "GCSStateRepo")
			return &domain.ExecutionState{
				LastExecutionTime: time.Time{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read state from GCS: %w", err)
	}
	defer reader.Close()

	// Read and parse JSON
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read state data: %w", err)
	}

	var stateData struct {
		LastExecutionTime time.Time `json:"last_execution_time"`
	}

	if err := json.Unmarshal(data, &stateData); err != nil {
		return nil, fmt.Errorf("failed to parse state JSON: %w", err)
	}

	r.logger.Log(domain.INFO, fmt.Sprintf("Retrieved state: last execution at %s", stateData.LastExecutionTime.Format(time.RFC3339)), "GCSStateRepo")

	return &domain.ExecutionState{
		LastExecutionTime: stateData.LastExecutionTime,
	}, nil
}

// SaveExecution saves the execution state to Cloud Storage
func (r *GCSStateRepository) SaveExecution(state *domain.ExecutionState) error {
	r.logger.Log(domain.INFO, fmt.Sprintf("Saving state to gs://%s/%s", r.bucketName, r.objectName), "GCSStateRepo")

	// Prepare JSON data
	stateData := struct {
		LastExecutionTime time.Time `json:"last_execution_time"`
	}{
		LastExecutionTime: state.LastExecutionTime,
	}

	jsonData, err := json.MarshalIndent(stateData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state to JSON: %w", err)
	}

	// Write to Cloud Storage
	bucket := r.client.Bucket(r.bucketName)
	obj := bucket.Object(r.objectName)

	writer := obj.NewWriter(r.ctx)
	writer.ContentType = "application/json"
	writer.Metadata = map[string]string{
		"updated_at": time.Now().Format(time.RFC3339),
	}

	if _, err := writer.Write(jsonData); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write state to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finalize GCS write: %w", err)
	}

	r.logger.Log(domain.INFO, fmt.Sprintf("State saved successfully: %s", state.LastExecutionTime.Format(time.RFC3339)), "GCSStateRepo")

	return nil
}
