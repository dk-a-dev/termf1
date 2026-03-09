package config

import "os"

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	GroqAPIKey  string
	GroqModel   string
	RefreshRate int // seconds between live-data refreshes
}

func Load() *Config {
	return &Config{
		GroqAPIKey:  os.Getenv("GROQ_API_KEY"),
		GroqModel:   getEnv("GROQ_MODEL", "compound-beta"),
		RefreshRate: getEnvInt("REFRESH_RATE", 5),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i := parseInt(v); i != 0 {
			return i
		}
	}
	return fallback
}
