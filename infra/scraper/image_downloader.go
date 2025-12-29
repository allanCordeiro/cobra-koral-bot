package scraper

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/allan/cobra-coral/internal/domain"
)

const (
	weatherMapURL = "https://www.saisp.br/cgesp/mapa_sp_geoserver.png"
	downloadTimeout = 30 * time.Second
)

// ImageDownloader interface for downloading images
type ImageDownloader interface {
	Download(url string) ([]byte, error)
}

// HTTPImageDownloader implements ImageDownloader using HTTP client
type HTTPImageDownloader struct {
	httpClient *http.Client
	logger     domain.Logger
}

// NewHTTPImageDownloader creates a new image downloader instance
func NewHTTPImageDownloader(logger domain.Logger) *HTTPImageDownloader {
	return &HTTPImageDownloader{
		httpClient: &http.Client{
			Timeout: downloadTimeout,
		},
		logger: logger,
	}
}

// Download downloads an image from the given URL and returns the bytes
func (d *HTTPImageDownloader) Download(url string) ([]byte, error) {
	d.logger.Log(domain.INFO, fmt.Sprintf("Downloading image from: %s", url), "HTTPImageDownloader")

	resp, err := d.httpClient.Get(url)
	if err != nil {
		d.logger.Log(domain.ERROR, "Failed to download image: "+err.Error(), "HTTPImageDownloader")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		d.logger.Log(domain.ERROR, err.Error(), "HTTPImageDownloader")
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		d.logger.Log(domain.ERROR, "Failed to read image data: "+err.Error(), "HTTPImageDownloader")
		return nil, err
	}

	d.logger.Log(domain.INFO, fmt.Sprintf("Successfully downloaded image (%d bytes)", len(data)), "HTTPImageDownloader")
	return data, nil
}

// DownloadWeatherMap is a convenience method to download the weather map
func (d *HTTPImageDownloader) DownloadWeatherMap() ([]byte, error) {
	return d.Download(weatherMapURL)
}
