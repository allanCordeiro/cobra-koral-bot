package domain

import "time"

// ExecutionState represents the last execution timestamp
type ExecutionState struct {
	LastExecutionTime time.Time
}

// StateRepository defines the interface for persisting execution state
type StateRepository interface {
	GetLastExecution() (*ExecutionState, error)
	SaveExecution(state *ExecutionState) error
}
