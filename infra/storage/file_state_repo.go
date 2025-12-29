package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/allan/cobra-coral/internal/domain"
)

const defaultStatePath = "./data/state.json"

// FileStateRepository implements domain.StateRepository using a JSON file
type FileStateRepository struct {
	filePath string
	logger   domain.Logger
}

// NewFileStateRepository creates a new file-based state repository
func NewFileStateRepository(logger domain.Logger, customPath string) (*FileStateRepository, error) {
	path := customPath
	if path == "" {
		path = defaultStatePath
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	repo := &FileStateRepository{
		filePath: path,
		logger:   logger,
	}

	// Initialize file if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		initialState := &domain.ExecutionState{
			LastExecutionTime: time.Time{}, // Zero time
		}
		if err := repo.SaveExecution(initialState); err != nil {
			return nil, err
		}
		logger.Log(domain.INFO, "Initialized state file with zero time", "FileStateRepository")
	}

	return repo, nil
}

type stateFile struct {
	LastExecutionTime time.Time `json:"last_execution_time"`
}

// GetLastExecution retrieves the last execution state from the file
func (r *FileStateRepository) GetLastExecution() (*domain.ExecutionState, error) {
	r.logger.Log(domain.INFO, "Reading execution state from file", "FileStateRepository")

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		r.logger.Log(domain.ERROR, "Failed to read state file: "+err.Error(), "FileStateRepository")
		return nil, err
	}

	var sf stateFile
	if err := json.Unmarshal(data, &sf); err != nil {
		r.logger.Log(domain.ERROR, "Failed to unmarshal state file: "+err.Error(), "FileStateRepository")
		return nil, err
	}

	state := &domain.ExecutionState{
		LastExecutionTime: sf.LastExecutionTime,
	}

	r.logger.Log(domain.INFO, "Successfully retrieved execution state", "FileStateRepository")
	return state, nil
}

// SaveExecution saves the execution state to the file atomically
func (r *FileStateRepository) SaveExecution(state *domain.ExecutionState) error {
	if state == nil {
		return errors.New("state cannot be nil")
	}

	r.logger.Log(domain.INFO, "Saving execution state to file", "FileStateRepository")

	sf := stateFile{
		LastExecutionTime: state.LastExecutionTime,
	}

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		r.logger.Log(domain.ERROR, "Failed to marshal state: "+err.Error(), "FileStateRepository")
		return err
	}

	// Atomic write using temp file + rename
	tempPath := r.filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		r.logger.Log(domain.ERROR, "Failed to write temp file: "+err.Error(), "FileStateRepository")
		return err
	}

	if err := os.Rename(tempPath, r.filePath); err != nil {
		r.logger.Log(domain.ERROR, "Failed to rename temp file: "+err.Error(), "FileStateRepository")
		os.Remove(tempPath) // Clean up temp file
		return err
	}

	r.logger.Log(domain.INFO, "Successfully saved execution state", "FileStateRepository")
	return nil
}
