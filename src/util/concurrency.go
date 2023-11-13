package util

const (
	defaultTotalWorkers = 650
)

type ConcurrencyConfig struct {
	TotalWorkers uint
}

func SetupDefaultConcurrency() *ConcurrencyConfig {
	return SetupConcurrency(defaultTotalWorkers)
}

// Returns a configuration struct to be used for channel initialization, default
// values are used if no overrides are passed in
func SetupConcurrency(customWorkerCount uint) *ConcurrencyConfig {
	cs := ConcurrencyConfig{
		TotalWorkers: customWorkerCount,
	}

	return &cs
}
