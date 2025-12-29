package telegram

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/allan/cobra-coral/internal/domain"
)

// Bot implements domain.NotificationService using Telegram
type Bot struct {
	api    *tgbotapi.BotAPI
	userID int64
	logger domain.Logger
}

// NewBot creates a new Telegram bot instance
func NewBot(token string, userIDStr string, logger domain.Logger) (*Bot, error) {
	if token == "" {
		return nil, errors.New("bot token is required")
	}
	if userIDStr == "" {
		return nil, errors.New("user ID is required")
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Log(domain.ERROR, "Failed to create Telegram bot: "+err.Error(), "TelegramBot")
		return nil, err
	}

	logger.Log(domain.INFO, fmt.Sprintf("Telegram bot authorized as: %s", api.Self.UserName), "TelegramBot")

	return &Bot{
		api:    api,
		userID: userID,
		logger: logger,
	}, nil
}

// SendMessage sends a notification message via Telegram
func (b *Bot) SendMessage(msg *domain.NotificationMessage) error {
	if msg == nil {
		return errors.New("message cannot be nil")
	}

	b.logger.Log(domain.INFO, "Sending message to Telegram", "TelegramBot")

	// If there's image data, send as photo with caption
	if len(msg.ImageData) > 0 {
		return b.sendPhotoWithCaption(msg.Text, msg.ImageData)
	}

	// Otherwise, send as text message
	return b.sendTextMessage(msg.Text)
}

// SendErrorAlert sends an error notification to Telegram
func (b *Bot) SendErrorAlert(errorMsg string) error {
	b.logger.Log(domain.WARN, "Sending error alert to Telegram", "TelegramBot")

	fullMessage := fmt.Sprintf("⚠️ *Weather Worker Error*\n\n%s", errorMsg)
	return b.sendTextMessage(fullMessage)
}

// sendTextMessage sends a plain text message
func (b *Bot) sendTextMessage(text string) error {
	msg := tgbotapi.NewMessage(b.userID, text)
	msg.ParseMode = "Markdown"

	b.logger.Log(domain.INFO, "Sending text message", "TelegramBot")

	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Log(domain.ERROR, "Failed to send text message: "+err.Error(), "TelegramBot")
		return err
	}

	b.logger.Log(domain.INFO, "Text message sent successfully", "TelegramBot")
	return nil
}

// sendPhotoWithCaption sends a photo with a caption
func (b *Bot) sendPhotoWithCaption(caption string, imageData []byte) error {
	b.logger.Log(domain.INFO, fmt.Sprintf("Sending photo message (%d bytes)", len(imageData)), "TelegramBot")

	// Create FileBytes from image data
	fileBytes := tgbotapi.FileBytes{
		Name:  "weather_map.png",
		Bytes: imageData,
	}

	msg := tgbotapi.NewPhoto(b.userID, fileBytes)
	msg.Caption = caption
	msg.ParseMode = "Markdown"

	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Log(domain.ERROR, "Failed to send photo: "+err.Error(), "TelegramBot")

		// Fallback: try sending text only if photo fails
		b.logger.Log(domain.WARN, "Attempting to send text-only message as fallback", "TelegramBot")
		return b.sendTextMessage(caption)
	}

	b.logger.Log(domain.INFO, "Photo message sent successfully", "TelegramBot")
	return nil
}

// sendPhotoReader is an alternative method using io.Reader
func (b *Bot) sendPhotoReader(caption string, imageData []byte) error {
	reader := bytes.NewReader(imageData)

	fileReader := tgbotapi.FileReader{
		Name:   "weather_map.png",
		Reader: reader,
	}

	msg := tgbotapi.NewPhoto(b.userID, fileReader)
	msg.Caption = caption
	msg.ParseMode = "Markdown"

	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Log(domain.ERROR, "Failed to send photo via reader: "+err.Error(), "TelegramBot")
		return err
	}

	return nil
}
