package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/allan/cobra-coral/internal/domain"
)

func TestJSONLogger_Log(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewJSONLogger()
	logger.Log(domain.INFO, "test message", "TestDomain")

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Parse JSON
	var entry logEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Verify fields
	if entry.Type != "INFO" {
		t.Errorf("Expected type INFO, got %s", entry.Type)
	}
	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", entry.Message)
	}
	if entry.Domain != "TestDomain" {
		t.Errorf("Expected domain 'TestDomain', got %s", entry.Domain)
	}
	if entry.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestJSONLogger_DifferentLogTypes(t *testing.T) {
	tests := []struct {
		name    string
		logType domain.LogType
		want    string
	}{
		{"INFO log", domain.INFO, "INFO"},
		{"ERROR log", domain.ERROR, "ERROR"},
		{"WARN log", domain.WARN, "WARN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logger := NewJSONLogger()
			logger.Log(tt.logType, "test", "test")

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)

			var entry logEntry
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			if entry.Type != tt.want {
				t.Errorf("Expected type %s, got %s", tt.want, entry.Type)
			}
		})
	}
}
