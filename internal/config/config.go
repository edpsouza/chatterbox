package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port      string
	Database  string
	JWTSecret string
	Debug     bool
}

func Load() Config {
	return Config{
		Port:      getEnv("PORT", "8080"),
		Database:  getEnv("DATABASE_URL", "chatterbox.db"),
		JWTSecret: getEnv("JWT_SECRET", "your_jwt_secret_here"),
		Debug:     getEnvBool("DEBUG", false),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return b
}
