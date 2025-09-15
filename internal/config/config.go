package config

// Config stores the application configuration.
type Config struct {
	PaperlessURL string
	APIKey       string
}

// Load loads the configuration from a file or environment variables.
func Load() (*Config, error) {
	// In a real application, you would load this from a file (e.g., YAML, JSON)
	// or from environment variables.
	return &Config{
		PaperlessURL: "http://localhost:8000",
		APIKey:       "your-api-key",
	}, nil
}
