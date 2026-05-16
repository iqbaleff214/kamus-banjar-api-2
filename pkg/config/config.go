package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string

	RedisAddr string
	RedisPass string

	JWTSecret string

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	OpenRouterAPIKey  string
	OpenRouterModel   string
	OpenRouterBaseURL string
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName,
	)
}

func Load() (*Config, error) {
	required := []string{
		"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASS",
		"REDIS_ADDR",
		"JWT_SECRET",
	}
	for _, key := range required {
		if os.Getenv(key) == "" {
			return nil, fmt.Errorf("missing required env var: %s", key)
		}
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		AppPort: getEnv("APP_PORT", "3000"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
		DBUser: os.Getenv("DB_USER"),
		DBPass: os.Getenv("DB_PASS"),

		RedisAddr: os.Getenv("REDIS_ADDR"),
		RedisPass: os.Getenv("REDIS_PASS"),

		JWTSecret: os.Getenv("JWT_SECRET"),

		SMTPHost: getEnv("SMTP_HOST", ""),
		SMTPPort: smtpPort,
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		SMTPFrom: getEnv("SMTP_FROM", ""),

		OpenRouterAPIKey:  getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:   getEnv("OPENROUTER_MODEL", "mistralai/mistral-7b-instruct:free"),
		OpenRouterBaseURL: getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
