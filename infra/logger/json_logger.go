package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/allan/cobra-coral/internal/domain"
)

// JSONLogger implements domain.Logger with structured JSON output
type JSONLogger struct{}

// NewJSONLogger creates a new JSON logger instance
func NewJSONLogger() *JSONLogger {
	return &JSONLogger{}
}

// logEntry represents the structure of a log entry
type logEntry struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Domain    string    `json:"domain"`
	Timestamp time.Time `json:"timestamp"`
}

// Log writes a structured JSON log entry to stdout
func (l *JSONLogger) Log(logType domain.LogType, message string, domainName string) {
	entry := logEntry{
		Type:      string(logType),
		Message:   message,
		Domain:    domainName,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to plain text if JSON marshaling fails
		fmt.Fprintf(os.Stdout, `{"type":"ERROR","message":"Failed to marshal log entry: %s","domain":"logger"}`+"\n", err)
		return
	}

	fmt.Fprintln(os.Stdout, string(data))
}
