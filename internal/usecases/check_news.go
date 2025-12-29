package usecases

import (
	"fmt"
	"strings"

	"github.com/allan/cobra-coral/internal/domain"
)

// Alert keywords in Portuguese
var alertKeywords = []string{
	"atenção",
	"alerta",
	"chuva",
	"alagamento",
	"temporal",
	"enchente",
	"inundação",
}

// CheckNewsUseCase analyzes news to determine if an alert should be sent
type CheckNewsUseCase struct {
	newsRepo  domain.NewsRepository
	stateRepo domain.StateRepository
	logger    domain.Logger
}

// NewCheckNewsUseCase creates a new CheckNewsUseCase instance
func NewCheckNewsUseCase(
	newsRepo domain.NewsRepository,
	stateRepo domain.StateRepository,
	logger domain.Logger,
) *CheckNewsUseCase {
	return &CheckNewsUseCase{
		newsRepo:  newsRepo,
		stateRepo: stateRepo,
		logger:    logger,
	}
}

// Execute checks if the news warrants an alert
// Returns: (shouldAlert bool, news *domain.News, error)
func (uc *CheckNewsUseCase) Execute() (bool, *domain.News, error) {
	uc.logger.Log(domain.INFO, "Starting news check", "CheckNewsUseCase")

	// Get last execution time
	state, err := uc.stateRepo.GetLastExecution()
	if err != nil {
		uc.logger.Log(domain.ERROR, "Failed to get last execution state: "+err.Error(), "CheckNewsUseCase")
		return false, nil, err
	}

	uc.logger.Log(domain.INFO, fmt.Sprintf("Last execution time: %s", state.LastExecutionTime), "CheckNewsUseCase")

	// Fetch latest news
	news, err := uc.newsRepo.FetchLatestNews()
	if err != nil {
		uc.logger.Log(domain.ERROR, "Failed to fetch latest news: "+err.Error(), "CheckNewsUseCase")
		return false, nil, err
	}

	uc.logger.Log(domain.INFO, fmt.Sprintf("Fetched news: %s (published at %s)", news.Title, news.PublishedAt), "CheckNewsUseCase")

	// Check if news is newer than last execution
	if !news.PublishedAt.After(state.LastExecutionTime) {
		uc.logger.Log(domain.INFO, "News is not newer than last execution, skipping alert", "CheckNewsUseCase")
		return false, news, nil
	}

	uc.logger.Log(domain.INFO, "News is newer than last execution, checking keywords", "CheckNewsUseCase")

	// Check for alert keywords
	if !uc.containsAlertKeywords(news) {
		uc.logger.Log(domain.INFO, "News does not contain alert keywords, skipping alert", "CheckNewsUseCase")
		return false, news, nil
	}

	uc.logger.Log(domain.INFO, "News contains alert keywords, alert should be sent", "CheckNewsUseCase")
	return true, news, nil
}

// containsAlertKeywords checks if the news contains any alert keywords
func (uc *CheckNewsUseCase) containsAlertKeywords(news *domain.News) bool {
	// Normalize text to lowercase for comparison
	titleLower := strings.ToLower(news.Title)
	contentLower := strings.ToLower(news.Content)

	for _, keyword := range alertKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(contentLower, keyword) {
			uc.logger.Log(domain.INFO, fmt.Sprintf("Found alert keyword: %s", keyword), "CheckNewsUseCase")
			return true
		}
	}

	return false
}
