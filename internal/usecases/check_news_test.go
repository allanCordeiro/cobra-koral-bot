package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/allan/cobra-coral/internal/domain"
)

// Mock logger
type mockLogger struct{}

func (m *mockLogger) Log(logType domain.LogType, message string, domainName string) {}

// Mock news repository
type mockNewsRepository struct {
	news *domain.News
	err  error
}

func (m *mockNewsRepository) FetchLatestNews() (*domain.News, error) {
	return m.news, m.err
}

// Mock state repository
type mockStateRepository struct {
	state *domain.ExecutionState
	err   error
}

func (m *mockStateRepository) GetLastExecution() (*domain.ExecutionState, error) {
	return m.state, m.err
}

func (m *mockStateRepository) SaveExecution(state *domain.ExecutionState) error {
	m.state = state
	return nil
}

func TestCheckNewsUseCase_NewerNewsWithKeywords(t *testing.T) {
	logger := &mockLogger{}

	lastExecution := time.Now().Add(-1 * time.Hour)
	stateRepo := &mockStateRepository{
		state: &domain.ExecutionState{
			LastExecutionTime: lastExecution,
		},
	}

	newsRepo := &mockNewsRepository{
		news: &domain.News{
			Title:       "Alerta de chuva forte",
			PublishedAt: time.Now(),
			Content:     "Previsão de temporal",
		},
	}

	uc := NewCheckNewsUseCase(newsRepo, stateRepo, logger)
	shouldAlert, news, err := uc.Execute()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !shouldAlert {
		t.Error("Expected shouldAlert to be true")
	}
	if news == nil {
		t.Error("Expected news to be returned")
	}
}

func TestCheckNewsUseCase_OlderNews(t *testing.T) {
	logger := &mockLogger{}

	lastExecution := time.Now()
	stateRepo := &mockStateRepository{
		state: &domain.ExecutionState{
			LastExecutionTime: lastExecution,
		},
	}

	newsRepo := &mockNewsRepository{
		news: &domain.News{
			Title:       "Alerta de chuva",
			PublishedAt: lastExecution.Add(-1 * time.Hour), // Older
			Content:     "Old news",
		},
	}

	uc := NewCheckNewsUseCase(newsRepo, stateRepo, logger)
	shouldAlert, _, err := uc.Execute()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if shouldAlert {
		t.Error("Expected shouldAlert to be false for older news")
	}
}

func TestCheckNewsUseCase_NoKeywords(t *testing.T) {
	logger := &mockLogger{}

	lastExecution := time.Now().Add(-1 * time.Hour)
	stateRepo := &mockStateRepository{
		state: &domain.ExecutionState{
			LastExecutionTime: lastExecution,
		},
	}

	newsRepo := &mockNewsRepository{
		news: &domain.News{
			Title:       "Sol durante o dia",
			PublishedAt: time.Now(),
			Content:     "Tempo bom",
		},
	}

	uc := NewCheckNewsUseCase(newsRepo, stateRepo, logger)
	shouldAlert, _, err := uc.Execute()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if shouldAlert {
		t.Error("Expected shouldAlert to be false when no keywords match")
	}
}

func TestCheckNewsUseCase_FetchError(t *testing.T) {
	logger := &mockLogger{}

	stateRepo := &mockStateRepository{
		state: &domain.ExecutionState{
			LastExecutionTime: time.Now(),
		},
	}

	newsRepo := &mockNewsRepository{
		err: errors.New("network error"),
	}

	uc := NewCheckNewsUseCase(newsRepo, stateRepo, logger)
	_, _, err := uc.Execute()

	if err == nil {
		t.Error("Expected error from news repository")
	}
}

func TestCheckNewsUseCase_StateError(t *testing.T) {
	logger := &mockLogger{}

	stateRepo := &mockStateRepository{
		err: errors.New("state error"),
	}

	newsRepo := &mockNewsRepository{
		news: &domain.News{},
	}

	uc := NewCheckNewsUseCase(newsRepo, stateRepo, logger)
	_, _, err := uc.Execute()

	if err == nil {
		t.Error("Expected error from state repository")
	}
}

func TestCheckNewsUseCase_KeywordDetection(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		content     string
		shouldMatch bool
	}{
		{"Atenção keyword", "Atenção para chuvas", "", true},
		{"Alerta keyword", "Alerta de temporal", "", true},
		{"Chuva keyword", "Previsão de chuva", "", true},
		{"Alagamento keyword", "Risco de alagamento", "", true},
		{"Temporal keyword", "Temporal se aproxima", "", true},
		{"Enchente keyword", "Enchente prevista", "", true},
		{"Inundação keyword", "Inundação em áreas baixas", "", true},
		{"No keywords", "Sol durante o dia", "Tempo bom", false},
		{"Keyword in content", "Previsão", "Risco de chuva forte", true},
		{"Case insensitive", "ALERTA DE CHUVA", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &mockLogger{}
			stateRepo := &mockStateRepository{
				state: &domain.ExecutionState{
					LastExecutionTime: time.Now().Add(-1 * time.Hour),
				},
			}

			newsRepo := &mockNewsRepository{
				news: &domain.News{
					Title:       tt.title,
					PublishedAt: time.Now(),
					Content:     tt.content,
				},
			}

			uc := NewCheckNewsUseCase(newsRepo, stateRepo, logger)
			shouldAlert, _, _ := uc.Execute()

			if shouldAlert != tt.shouldMatch {
				t.Errorf("Expected shouldAlert=%v, got %v", tt.shouldMatch, shouldAlert)
			}
		})
	}
}
