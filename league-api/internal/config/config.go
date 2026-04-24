package config

import "os"

type Config struct {
	DatabaseURL          string
	Port                 string
	JWTSecret            string
	GoogleClientID       string
	GoogleClientSecret   string
	FacebookClientID     string
	FacebookClientSecret string
	AppleClientID        string
	AppleTeamID          string
	AppleKeyID           string
	ApplePrivateKey      string
	FrontendURL          string
}

func Load() Config {
	return Config{
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://test:test@localhost:5432/test?sslmode=disable"),
		Port:                 getEnv("PORT", "8080"),
		JWTSecret:            getEnv("JWT_SECRET", "change-me"),
		GoogleClientID:       os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		FacebookClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
		FacebookClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
		AppleClientID:        os.Getenv("APPLE_CLIENT_ID"),
		AppleTeamID:          os.Getenv("APPLE_TEAM_ID"),
		AppleKeyID:           os.Getenv("APPLE_KEY_ID"),
		ApplePrivateKey:      os.Getenv("APPLE_PRIVATE_KEY"),
		FrontendURL:          getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
