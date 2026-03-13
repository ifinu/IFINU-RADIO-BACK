package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	ServerPort string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Streaming
	StreamBufferSize    int
	StreamReadTimeout   time.Duration
	StreamWriteTimeout  time.Duration
	StreamRetryAttempts int
	StreamRetryDelay    time.Duration

	// HTTP Client
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration

	// Sync
	SyncInterval time.Duration
}

func Load() *Config {
	return &Config{
		ServerPort:          getEnv("PORT", "8080"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPassword:          getEnv("DB_PASSWORD", "postgres"),
		DBName:              getEnv("DB_NAME", "radio_db"),
		StreamBufferSize:    getEnvInt("STREAM_BUFFER_SIZE", 65536), // 64KB
		StreamReadTimeout:   getEnvDuration("STREAM_READ_TIMEOUT", 30*time.Second),
		StreamWriteTimeout:  getEnvDuration("STREAM_WRITE_TIMEOUT", 30*time.Second),
		StreamRetryAttempts: getEnvInt("STREAM_RETRY_ATTEMPTS", 3),
		StreamRetryDelay:    getEnvDuration("STREAM_RETRY_DELAY", 2*time.Second),
		MaxIdleConns:        getEnvInt("MAX_IDLE_CONNS", 100),
		MaxIdleConnsPerHost: getEnvInt("MAX_IDLE_CONNS_PER_HOST", 50),
		IdleConnTimeout:     getEnvDuration("IDLE_CONN_TIMEOUT", 90*time.Second),
		SyncInterval:        getEnvDuration("SYNC_INTERVAL", 6*time.Hour),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
