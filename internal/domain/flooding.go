package domain

// FloodingPoint represents a flooding location
type FloodingPoint struct {
	Location  string
	TimeRange string
	Direction string
	Reference string
	Zone      string
}

// FloodingRepository defines the interface for fetching flooding data
type FloodingRepository interface {
	GetActiveCount() (int, error)
	FetchFloodingPoints(date string) ([]FloodingPoint, error)
}
