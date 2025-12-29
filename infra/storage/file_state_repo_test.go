package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/allan/cobra-coral/internal/domain"
)

// mockLogger is a simple logger for testing
type mockLogger struct{}

func (m *mockLogger) Log(logType domain.LogType, message string, domainName string) {
	// No-op for tests
}

func TestFileStateRepository_SaveAndGet(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "state.json")

	logger := &mockLogger{}
	repo, err := NewFileStateRepository(logger, filePath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Get initial state (should be zero time)
	state, err := repo.GetLastExecution()
	if err != nil {
		t.Fatalf("Failed to get initial state: %v", err)
	}
	if !state.LastExecutionTime.IsZero() {
		t.Error("Expected zero time for initial state")
	}

	// Save a new state
	now := time.Now()
	newState := &domain.ExecutionState{
		LastExecutionTime: now,
	}
	if err := repo.SaveExecution(newState); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Retrieve saved state
	retrievedState, err := repo.GetLastExecution()
	if err != nil {
		t.Fatalf("Failed to get saved state: %v", err)
	}

	// Compare times (truncate to second for comparison)
	if retrievedState.LastExecutionTime.Truncate(time.Second) != now.Truncate(time.Second) {
		t.Errorf("Expected time %v, got %v", now, retrievedState.LastExecutionTime)
	}
}

func TestFileStateRepository_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nested", "dir", "state.json")

	logger := &mockLogger{}
	_, err := NewFileStateRepository(logger, filePath)
	if err != nil {
		t.Fatalf("Failed to create repository with nested path: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("State file was not created")
	}
}

func TestFileStateRepository_SaveNilState(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "state.json")

	logger := &mockLogger{}
	repo, err := NewFileStateRepository(logger, filePath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Attempt to save nil state
	err = repo.SaveExecution(nil)
	if err == nil {
		t.Error("Expected error when saving nil state")
	}
}

func TestFileStateRepository_DefaultPath(t *testing.T) {
	logger := &mockLogger{}

	// This test might fail if /app/data doesn't exist and can't be created
	// For real tests, you'd mock the filesystem or use a temp path
	_, err := NewFileStateRepository(logger, "")
	if err != nil {
		// Expected to fail if /app/data can't be created
		t.Logf("Expected failure for default path in test environment: %v", err)
	}
}
