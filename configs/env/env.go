package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port   string
	AppEnv string

	// MongoDB
	MongoURI    string
	MongoDBName string

	// Redis
	RedisURI      string
	RedisPassword string
	RedisUser     string
	RedisPrefix   string

	// Payment notifications
	PaymentReminderDays         int
	SMTPHost                    string
	SMTPPort                    int
	SMTPUsername                string
	SMTPPassword                string
	SMTPFrom                    string
	TelegramBotToken            string
	NotificationIntervalMinutes int
	GmailUsername               string
	GmailAppPassword            string
	GmailFrom                   string
	AuthSecret                  string
	AuthTokenHours              int

	// OpenTelemetry
	OTLPEndpoint string
	ServiceName  string
}

var config *Config

// Load loads environment variables from .env file
func Load() error {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Initialize config
	config = &Config{
		Port:                        getEnv("PORT", "8080"),
		AppEnv:                      getEnv("APP_ENV", "development"),
		MongoURI:                    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:                 getEnv("MONGO_DB_NAME", "dp-vc-webApp"),
		RedisURI:                    getEnv("REDIS_URI", "localhost:6379"),
		RedisPassword:               getEnv("REDIS_PASSWORD", ""),
		RedisUser:                   getEnv("REDIS_USER", ""),
		RedisPrefix:                 getEnv("REDIS_PREFIX", "dp-vc-webApp"),
		PaymentReminderDays:         GetEnvInt("PAYMENT_REMINDER_DAYS", 1),
		SMTPHost:                    getEnv("SMTP_HOST", ""),
		SMTPPort:                    GetEnvInt("SMTP_PORT", 587),
		SMTPUsername:                getEnv("SMTP_USERNAME", ""),
		SMTPPassword:                getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:                    getEnv("SMTP_FROM", ""),
		TelegramBotToken:            getEnv("TELEGRAM_BOT_TOKEN", ""),
		NotificationIntervalMinutes: GetEnvInt("NOTIFICATION_INTERVAL_MINUTES", 30),
		GmailUsername:               getEnv("GMAIL_USERNAME", ""),
		GmailAppPassword:            getEnv("GMAIL_APP_PASSWORD", ""),
		GmailFrom:                   getEnv("GMAIL_FROM", ""),
		AuthSecret:                  getEnv("AUTH_SECRET", "change-this-secret-in-production"),
		AuthTokenHours:              GetEnvInt("AUTH_TOKEN_HOURS", 24),
		OTLPEndpoint:                getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		ServiceName:                 getEnv("OTEL_SERVICE_NAME", "dp-vc-webApp"),
	}

	return nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	return config
}

// GetEnv returns environment variable value or default
func GetEnv(key, defaultValue string) string {
	return getEnv(key, defaultValue)
}

// GetEnvInt returns environment variable as int or default
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetEnvBool returns environment variable as bool or default
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func IsDevelopment() bool {
	return config != nil && config.AppEnv == "development"
}

// IsProduction returns true if running in production mode
func IsProduction() bool {
	return config != nil && config.AppEnv == "production"
}

// IsStaging returns true if running in staging mode
func IsStaging() bool {
	return config != nil && config.AppEnv == "staging"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
