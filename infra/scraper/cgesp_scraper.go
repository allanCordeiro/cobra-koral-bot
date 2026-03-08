package scraper

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/allan/cobra-coral/internal/domain"
	"golang.org/x/net/html"
)

const (
	baseURL          = "https://www.cgesp.org/v3"
	indexURL         = baseURL + "/index.jsp"
	floodingURL      = baseURL + "/alagamentos.jsp"
	dateTimeLayout   = "02/01/2006 15:04"
	defaultHTTPTimeout = 30 * time.Second
)

// CGESPScraper implements NewsRepository and FloodingRepository
type CGESPScraper struct {
	httpClient *http.Client
	logger     domain.Logger
}

// NewCGESPScraper creates a new scraper instance
func NewCGESPScraper(logger domain.Logger) *CGESPScraper {
	return &CGESPScraper{
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		logger: logger,
	}
}

// FetchLatestNews scrapes and returns the latest news from the main page
func (s *CGESPScraper) FetchLatestNews() (*domain.News, error) {
	s.logger.Log(domain.INFO, "Fetching latest news from CGE São Paulo", "CGESPScraper")

	resp, err := s.httpClient.Get(indexURL)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to fetch index page: "+err.Error(), "CGESPScraper")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		s.logger.Log(domain.ERROR, err.Error(), "CGESPScraper")
		return nil, err
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to parse HTML: "+err.Error(), "CGESPScraper")
		return nil, err
	}

	news, err := s.extractFirstNews(doc)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to extract news: "+err.Error(), "CGESPScraper")
		return nil, err
	}

	s.logger.Log(domain.INFO, fmt.Sprintf("Successfully extracted news: %s", news.Title), "CGESPScraper")
	return news, nil
}

// GetActiveCount returns the number of active flooding points
func (s *CGESPScraper) GetActiveCount() (int, error) {
	s.logger.Log(domain.INFO, "Fetching active flooding count", "CGESPScraper")

	resp, err := s.httpClient.Get(indexURL)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to fetch index page: "+err.Error(), "CGESPScraper")
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		s.logger.Log(domain.ERROR, err.Error(), "CGESPScraper")
		return 0, err
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to parse HTML: "+err.Error(), "CGESPScraper")
		return 0, err
	}

	count, err := s.extractFloodingCount(doc)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to extract flooding count: "+err.Error(), "CGESPScraper")
		return 0, err
	}

	s.logger.Log(domain.INFO, fmt.Sprintf("Active flooding points: %d", count), "CGESPScraper")
	return count, nil
}

// FetchFloodingPoints fetches flooding points for a specific date
func (s *CGESPScraper) FetchFloodingPoints(date string) ([]domain.FloodingPoint, error) {
	s.logger.Log(domain.INFO, fmt.Sprintf("Fetching flooding points for date: %s", date), "CGESPScraper")

	url := fmt.Sprintf("%s?dataBusca=%s&enviaBusca=Buscar", floodingURL, date)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to fetch flooding page: "+err.Error(), "CGESPScraper")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		s.logger.Log(domain.ERROR, err.Error(), "CGESPScraper")
		return nil, err
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to parse HTML: "+err.Error(), "CGESPScraper")
		return nil, err
	}

	points, err := s.extractFloodingPoints(doc)
	if err != nil {
		s.logger.Log(domain.ERROR, "Failed to extract flooding points: "+err.Error(), "CGESPScraper")
		return nil, err
	}

	s.logger.Log(domain.INFO, fmt.Sprintf("Successfully extracted %d flooding points", len(points)), "CGESPScraper")
	return points, nil
}

// extractFirstNews finds the first news item in the HTML document
func (s *CGESPScraper) extractFirstNews(n *html.Node) (*domain.News, error) {
	var newsAnchor *html.Node
	var findNews func(*html.Node)
	findNews = func(node *html.Node) {
		if newsAnchor != nil {
			return
		}
		if node.Type == html.ElementNode && node.Data == "a" {
			// Check if this anchor has href with "noticias.jsp?id=" (actual news, not menu link)
			for _, attr := range node.Attr {
				if attr.Key == "href" && strings.Contains(attr.Val, "noticias.jsp?id=") {
					newsAnchor = node
					return
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			findNews(c)
		}
	}
	findNews(n)

	if newsAnchor == nil {
		return nil, errors.New("no news anchor found in HTML")
	}

	// Extract title from h2 or h3 and datetime from p
	var title string
	var dateTimeStr string

	var extractContent func(*html.Node)
	extractContent = func(node *html.Node) {
		if node.Type == html.ElementNode {
			// Extract title from h1, h2, or h3
			if node.Data == "h1" || node.Data == "h2" || node.Data == "h3" {
				text := strings.TrimSpace(getTextContent(node))

				// Check if this looks like a date/time (has / and :)
				if strings.Contains(text, "/") && strings.Contains(text, ":") {
					// This is the datetime
					if dateTimeStr == "" {
						dateTimeStr = text
					}
				} else if text != "" && title == "" {
					// This is the title (first non-date heading)
					title = text
				}
			}

			// Extract datetime from p tag (alternative location)
			if node.Data == "p" {
				text := strings.TrimSpace(getTextContent(node))
				if strings.Contains(text, "/") && strings.Contains(text, ":") && dateTimeStr == "" {
					dateTimeStr = text
				}
			}
		}

		// Recursively process children
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extractContent(c)
		}
	}

	// Start extraction from anchor's children
	for child := newsAnchor.FirstChild; child != nil; child = child.NextSibling {
		extractContent(child)
	}

	if title == "" || dateTimeStr == "" {
		return nil, fmt.Errorf("failed to extract title or datetime from news (title: '%s', datetime: '%s')", title, dateTimeStr)
	}

	// Parse datetime: "28/12/2025 21:00 - Domingo"
	publishedAt, err := s.parseDateTime(dateTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datetime: %w", err)
	}

	news := &domain.News{
		Title:       strings.TrimSpace(title),
		PublishedAt: publishedAt,
		Content:     strings.TrimSpace(title), // Using title as content for now
	}

	return news, nil
}

// extractFloodingCount extracts the number of active flooding points
func (s *CGESPScraper) extractFloodingCount(n *html.Node) (int, error) {
	var count int
	var found bool

	var findCount func(*html.Node)
	findCount = func(node *html.Node) {
		if found {
			return
		}

		// Check any element that might contain the text
		if node.Type == html.ElementNode {
			text := getTextContent(node)
			// Looking for "Pontos de Alagamento: X ativos"
			if strings.Contains(text, "Pontos de Alagamento") && strings.Contains(text, "ativos") {
				// Extract number using regex
				re := regexp.MustCompile(`(\d+)\s+ativos`)
				matches := re.FindStringSubmatch(text)
				if len(matches) >= 2 {
					parsedCount, err := strconv.Atoi(matches[1])
					if err == nil {
						count = parsedCount
						found = true
						return
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			findCount(c)
		}
	}
	findCount(n)

	if !found {
		return 0, errors.New("flooding count not found in HTML")
	}

	return count, nil
}

// extractFloodingPoints extracts flooding point details from the flooding page
func (s *CGESPScraper) extractFloodingPoints(n *html.Node) ([]domain.FloodingPoint, error) {
	var points []domain.FloodingPoint
	var currentZone string
	var currentBairro string

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		// Find zone headers (h1 with class "tit-bairros")
		if node.Type == html.ElementNode && node.Data == "h1" {
			if hasClass(node, "tit-bairros") {
				zoneText := getTextContent(node)
				if strings.Contains(zoneText, "Zona") {
					currentZone = strings.TrimSpace(zoneText)
				}
			}
		}

		// Find bairro (neighborhood) in table cells
		if node.Type == html.ElementNode && node.Data == "td" {
			if hasClass(node, "bairro") {
				bairroText := getTextContent(node)
				// Remove trailing whitespace and "pts" text
				bairroText = strings.Split(bairroText, "\n")[0]
				currentBairro = strings.TrimSpace(bairroText)
			}
		}

		// Find flooding points in div with class "ponto-de-alagamento"
		if node.Type == html.ElementNode && node.Data == "div" {
			if hasClass(node, "ponto-de-alagamento") && currentZone != "" {
				point := s.parseFloodingPointFromDiv(node, currentZone, currentBairro)
				if point != nil {
					points = append(points, *point)
				}
				return // Don't traverse children, we already processed them
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)

	return points, nil
}

// hasClass checks if an HTML node has a specific class
func hasClass(node *html.Node, className string) bool {
	for _, attr := range node.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, c := range classes {
				if c == className {
					return true
				}
			}
		}
	}
	return false
}

// parseFloodingPointFromDiv extracts flooding info from the new HTML structure
func (s *CGESPScraper) parseFloodingPointFromDiv(div *html.Node, zone, bairro string) *domain.FloodingPoint {
	var timeRange, location, direction, reference string
	var isActive bool

	var extractFromUL func(*html.Node)
	extractFromUL = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "li" {
			// Check if this is a status indicator (active/inactive)
			if hasClass(node, "ativo-transitavel") || hasClass(node, "ativo-intransitavel") {
				isActive = true
			}

			// Check for location and time info (class col-local)
			if hasClass(node, "col-local") {
				text := getTextContent(node)
				// Format can be:
				// "De 14:19 a \nAV JOSE PINHEIRO BORGES"
				// "De 14:14 a 14:32\nAV ITAQUERA"
				// "De 14:03 a AV MORVAN DIAS DE FIGUEIREDO" (without newline)
				lines := strings.Split(text, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "De ") {
						// Try to extract time and location from the same line
						// Pattern: "De HH:MM a [HH:MM ]LOCATION" or "De HH:MM a \nLOCATION"
						timePattern := regexp.MustCompile(`^(De \d{1,2}:\d{2} a(?: \d{1,2}:\d{2})?)(.*)$`)
						matches := timePattern.FindStringSubmatch(line)
						if len(matches) >= 3 {
							timeRange = strings.TrimSpace(matches[1])
							remaining := strings.TrimSpace(matches[2])
							if remaining != "" && location == "" {
								location = remaining
							}
						} else {
							timeRange = line
						}
					} else if line != "" && location == "" {
						location = line
					}
				}
			}

			// Check for direction and reference info
			if hasClass(node, "arial-descr-alag") && !hasClass(node, "col-local") {
				text := getTextContent(node)
				lines := strings.Split(text, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "Sentido:") {
						direction = strings.TrimSpace(strings.TrimPrefix(line, "Sentido:"))
					} else if strings.HasPrefix(line, "Referência:") {
						reference = strings.TrimSpace(strings.TrimPrefix(line, "Referência:"))
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extractFromUL(c)
		}
	}
	extractFromUL(div)

	// Only return active flooding points
	if !isActive || location == "" {
		return nil
	}

	// Include bairro in zone if available
	fullZone := zone
	if bairro != "" {
		fullZone = zone + " - " + bairro
	}

	return &domain.FloodingPoint{
		Location:  location,
		TimeRange: timeRange,
		Direction: direction,
		Reference: reference,
		Zone:      fullZone,
	}
}

// parseDateTime parses a datetime string in the format "28/12/2025 21:00 - Domingo"
func (s *CGESPScraper) parseDateTime(dateTimeStr string) (time.Time, error) {
	// Extract just the date and time part before the dash
	parts := strings.Split(dateTimeStr, " - ")
	if len(parts) == 0 {
		return time.Time{}, errors.New("invalid datetime format")
	}

	dateTimePart := strings.TrimSpace(parts[0])
	parsedTime, err := time.Parse(dateTimeLayout, dateTimePart)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

// getTextContent recursively extracts all text content from a node
func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	// Insert newline for <br> elements
	if n.Type == html.ElementNode && n.Data == "br" {
		return "\n"
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getTextContent(c)
	}
	return text
}
