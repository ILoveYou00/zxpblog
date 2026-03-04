package config

import (
	"os"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	AppPort      string
	GinMode      string
	AdminUsername string
	AdminPassword string
	SessionSecret string
	// AI 配置
	AIEnabled  bool
	AIApiKey   string
	AIApiUrl   string
	AIModel    string
	// OAuth 配置
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
}

func Load() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "blog"),
		AppPort:       getEnv("APP_PORT", "8080"),
		GinMode:       getEnv("GIN_MODE", "debug"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		SessionSecret: getEnv("SESSION_SECRET", "secret"),
		// AI 配置
		AIEnabled:  getEnv("AI_API_KEY", "") != "",
		AIApiKey:   getEnv("AI_API_KEY", ""),
		AIApiUrl:   getEnv("AI_API_URL", ""),
		AIModel:    getEnv("AI_MODEL", ""),
		// OAuth 配置
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}