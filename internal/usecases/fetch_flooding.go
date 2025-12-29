package usecases

import (
	"fmt"
	"net/url"
	"time"

	"github.com/allan/cobra-coral/internal/domain"
)

// FetchFloodingUseCase fetches flooding data if there are active points
type FetchFloodingUseCase struct {
	floodingRepo domain.FloodingRepository
	logger       domain.Logger
}

// NewFetchFloodingUseCase creates a new FetchFloodingUseCase instance
func NewFetchFloodingUseCase(
	floodingRepo domain.FloodingRepository,
	logger domain.Logger,
) *FetchFloodingUseCase {
	return &FetchFloodingUseCase{
		floodingRepo: floodingRepo,
		logger:       logger,
	}
}

// Execute fetches flooding points if any are active
// Returns: (hasFlooding bool, points []FloodingPoint, error)
func (uc *FetchFloodingUseCase) Execute() (bool, []domain.FloodingPoint, error) {
	uc.logger.Log(domain.INFO, "Starting flooding check", "FetchFloodingUseCase")

	// Get active flooding count
	count, err := uc.floodingRepo.GetActiveCount()
	if err != nil {
		uc.logger.Log(domain.ERROR, "Failed to get active flooding count: "+err.Error(), "FetchFloodingUseCase")
		return false, nil, err
	}

	uc.logger.Log(domain.INFO, fmt.Sprintf("Active flooding count: %d", count), "FetchFloodingUseCase")

	if count == 0 {
		uc.logger.Log(domain.INFO, "No active flooding points", "FetchFloodingUseCase")
		return false, nil, nil
	}

	// Fetch flooding points for today
	today := time.Now().Format("02/01/2006")
	encodedDate := url.QueryEscape(today)

	uc.logger.Log(domain.INFO, fmt.Sprintf("Fetching flooding points for date: %s", today), "FetchFloodingUseCase")

	points, err := uc.floodingRepo.FetchFloodingPoints(encodedDate)
	if err != nil {
		uc.logger.Log(domain.ERROR, "Failed to fetch flooding points: "+err.Error(), "FetchFloodingUseCase")
		return false, nil, err
	}

	uc.logger.Log(domain.INFO, fmt.Sprintf("Successfully fetched %d flooding points", len(points)), "FetchFloodingUseCase")
	return true, points, nil
}
