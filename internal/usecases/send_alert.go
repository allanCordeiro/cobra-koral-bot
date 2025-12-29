package usecases

import (
	"fmt"
	"strings"

	"github.com/allan/cobra-coral/internal/domain"
	"github.com/allan/cobra-coral/infra/scraper"
)

// SendAlertUseCase orchestrates sending alerts via Telegram
type SendAlertUseCase struct {
	notificationSvc domain.NotificationService
	imageDownloader scraper.ImageDownloader
	logger          domain.Logger
}

// NewSendAlertUseCase creates a new SendAlertUseCase instance
func NewSendAlertUseCase(
	notificationSvc domain.NotificationService,
	imageDownloader scraper.ImageDownloader,
	logger domain.Logger,
) *SendAlertUseCase {
	return &SendAlertUseCase{
		notificationSvc: notificationSvc,
		imageDownloader: imageDownloader,
		logger:          logger,
	}
}

// Execute builds and sends the alert message
func (uc *SendAlertUseCase) Execute(news *domain.News, floodingPoints []domain.FloodingPoint) error {
	uc.logger.Log(domain.INFO, "Building alert message", "SendAlertUseCase")

	// Build message text
	var messageBuilder strings.Builder

	messageBuilder.WriteString("🌦️ *Alerta Meteorológico - CGE São Paulo*\n\n")
	messageBuilder.WriteString(fmt.Sprintf("*%s*\n\n", news.Title))
	messageBuilder.WriteString(fmt.Sprintf("📅 Publicado em: %s\n\n", news.PublishedAt.Format("02/01/2006 15:04")))

	if news.Content != "" && news.Content != news.Title {
		messageBuilder.WriteString(fmt.Sprintf("%s\n\n", news.Content))
	}

	// Add flooding information if present
	if len(floodingPoints) > 0 {
		messageBuilder.WriteString(fmt.Sprintf("🚨 *Pontos de Alagamento Ativos: %d*\n\n", len(floodingPoints)))

		// Group by zone
		zones := make(map[string][]domain.FloodingPoint)
		for _, point := range floodingPoints {
			zones[point.Zone] = append(zones[point.Zone], point)
		}

		for zone, points := range zones {
			messageBuilder.WriteString(fmt.Sprintf("*%s* (%d pontos)\n", zone, len(points)))
			for _, point := range points {
				messageBuilder.WriteString(fmt.Sprintf("• %s\n", point.Location))
				if point.TimeRange != "" {
					messageBuilder.WriteString(fmt.Sprintf("  %s\n", point.TimeRange))
				}
				if point.Direction != "" {
					messageBuilder.WriteString(fmt.Sprintf("  Sentido: %s\n", point.Direction))
				}
				if point.Reference != "" {
					messageBuilder.WriteString(fmt.Sprintf("  Ref: %s\n", point.Reference))
				}
			}
			messageBuilder.WriteString("\n")
		}
	}

	messageText := messageBuilder.String()

	// Download weather map image
	uc.logger.Log(domain.INFO, "Downloading weather map image", "SendAlertUseCase")
	imageData, err := uc.imageDownloader.Download("https://www.saisp.br/cgesp/mapa_sp_geoserver.png")
	if err != nil {
		uc.logger.Log(domain.WARN, "Failed to download weather map, sending message without image: "+err.Error(), "SendAlertUseCase")
		// Continue without image
		imageData = nil
	}

	// Send notification
	msg := &domain.NotificationMessage{
		Text:      messageText,
		ImageData: imageData,
	}

	uc.logger.Log(domain.INFO, "Sending notification via Telegram", "SendAlertUseCase")
	if err := uc.notificationSvc.SendMessage(msg); err != nil {
		uc.logger.Log(domain.ERROR, "Failed to send notification: "+err.Error(), "SendAlertUseCase")
		return err
	}

	uc.logger.Log(domain.INFO, "Successfully sent alert notification", "SendAlertUseCase")
	return nil
}
