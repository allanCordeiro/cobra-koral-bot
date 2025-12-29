package domain

// LogType represents the severity level of a log message
type LogType string

const (
	INFO  LogType = "INFO"
	ERROR LogType = "ERROR"
	WARN  LogType = "WARN"
)

// Logger defines the interface for structured logging
type Logger interface {
	Log(logType LogType, message string, domain string)
}
