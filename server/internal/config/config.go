package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	FrontendURL       string
	RoundDefaultTime  int
	TopicDefaultTime  int
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:             getEnv("PORT", "8080"),
		FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:5173"),
		RoundDefaultTime: getEnvInt("ROUND_DEFAULT_TIME", 300),
		TopicDefaultTime: getEnvInt("TOPIC_DEFAULT_TIME", 60),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
