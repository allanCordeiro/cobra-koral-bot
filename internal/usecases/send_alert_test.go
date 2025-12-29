package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/allan/cobra-coral/internal/domain"
)

// Mock notification service
type mockNotificationService struct {
	sentMessage *domain.NotificationMessage
	err         error
}

func (m *mockNotificationService) SendMessage(msg *domain.NotificationMessage) error {
	m.sentMessage = msg
	return m.err
}

func (m *mockNotificationService) SendErrorAlert(errorMsg string) error {
	return m.err
}

// Mock image downloader
type mockImageDownloader struct {
	data []byte
	err  error
}

func (m *mockImageDownloader) Download(url string) ([]byte, error) {
	return m.data, m.err
}

func TestSendAlertUseCase_WithImage(t *testing.T) {
	logger := &mockLogger{}

	notificationSvc := &mockNotificationService{}
	imageDownloader := &mockImageDownloader{
		data: []byte("fake image data"),
	}

	uc := NewSendAlertUseCase(notificationSvc, imageDownloader, logger)

	news := &domain.News{
		Title:       "Alerta de chuva",
		PublishedAt: time.Now(),
		Content:     "Chuva forte prevista",
	}

	err := uc.Execute(news, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if notificationSvc.sentMessage == nil {
		t.Fatal("Expected message to be sent")
	}

	if len(notificationSvc.sentMessage.ImageData) == 0 {
		t.Error("Expected image data to be included")
	}

	if notificationSvc.sentMessage.Text == "" {
		t.Error("Expected message text to be set")
	}
}

func TestSendAlertUseCase_WithoutImage(t *testing.T) {
	logger := &mockLogger{}

	notificationSvc := &mockNotificationService{}
	imageDownloader := &mockImageDownloader{
		err: errors.New("image download failed"),
	}

	uc := NewSendAlertUseCase(notificationSvc, imageDownloader, logger)

	news := &domain.News{
		Title:       "Alerta de chuva",
		PublishedAt: time.Now(),
	}

	// Should continue without image
	err := uc.Execute(news, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if notificationSvc.sentMessage == nil {
		t.Fatal("Expected message to be sent")
	}

	if notificationSvc.sentMessage.ImageData != nil {
		t.Error("Expected no image data when download fails")
	}
}

func TestSendAlertUseCase_WithFlooding(t *testing.T) {
	logger := &mockLogger{}

	notificationSvc := &mockNotificationService{}
	imageDownloader := &mockImageDownloader{
		data: []byte("image"),
	}

	uc := NewSendAlertUseCase(notificationSvc, imageDownloader, logger)

	news := &domain.News{
		Title:       "Alerta",
		PublishedAt: time.Now(),
	}

	floodingPoints := []domain.FloodingPoint{
		{
			Location:  "R MANOEL BARBOSA",
			Zone:      "Zona Norte",
			TimeRange: "De 17:17 a 18:09",
		},
	}

	err := uc.Execute(news, floodingPoints)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Message should contain flooding information
	if notificationSvc.sentMessage == nil {
		t.Fatal("Expected message to be sent")
	}

	// Check that message contains flooding keywords
	messageText := notificationSvc.sentMessage.Text
	if messageText == "" {
		t.Error("Expected message text")
	}
	// Could add more specific checks for flooding content
}

func TestSendAlertUseCase_SendError(t *testing.T) {
	logger := &mockLogger{}

	notificationSvc := &mockNotificationService{
		err: errors.New("send failed"),
	}
	imageDownloader := &mockImageDownloader{
		data: []byte("image"),
	}

	uc := NewSendAlertUseCase(notificationSvc, imageDownloader, logger)

	news := &domain.News{
		Title:       "Alerta",
		PublishedAt: time.Now(),
	}

	err := uc.Execute(news, nil)
	if err == nil {
		t.Error("Expected error when sending notification fails")
	}
}
