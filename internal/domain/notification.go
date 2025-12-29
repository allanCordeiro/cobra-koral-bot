package domain

// NotificationMessage represents a message to be sent via Telegram
type NotificationMessage struct {
	Text      string
	ImageData []byte
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendMessage(msg *NotificationMessage) error
	SendErrorAlert(errorMsg string) error
}
