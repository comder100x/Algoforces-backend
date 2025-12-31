// configuration.go

package conf

import (
	"fmt"
	"os"
)

// Default configuration values
const (
	defaultJWT_SECRET     = "dev-secret-key-change-in-production"
	defaultDB_HOST        = "host.docker.internal"
	defaultDB_PORT        = "5432"
	defaultDB_USER        = "postgres"
	defaultDB_PASSWORD    = "postgres"
	defaultDB_NAME        = "algoforces"
	defaultDB_SSLMODE     = "disable"
	defaultREDIS_ADDR     = "localhost:6379"
	defaultJUDGE0_URL     = "http://localhost:2358"
	defaultJUDGE0_API_KEY = ""
)

// Configuration variables with defaults and environment overrides
var (
	JWT_SECRET     string
	DB_HOST        string
	DB_PORT        string
	DB_USER        string
	DB_PASSWORD    string
	DB_NAME        string
	DB_SSLMODE     string
	REDIS_URL      string
	JUDGE0_URL     string
	JUDGE0_API_KEY string
)

// init function runs when the package is imported
func init() {
	// Initialize with defaults first
	JWT_SECRET = defaultJWT_SECRET
	DB_HOST = defaultDB_HOST
	DB_PORT = defaultDB_PORT
	DB_USER = defaultDB_USER
	DB_PASSWORD = defaultDB_PASSWORD
	DB_NAME = defaultDB_NAME
	DB_SSLMODE = defaultDB_SSLMODE
	REDIS_URL = defaultREDIS_ADDR
	JUDGE0_URL = defaultJUDGE0_URL
	JUDGE0_API_KEY = defaultJUDGE0_API_KEY
	fmt.Println("db host", DB_HOST)

	// Override with environment variables if they exist
	if envValue := os.Getenv("JWT_SECRET"); envValue != "" {
		JWT_SECRET = envValue
	}
	if envValue := os.Getenv("DB_HOST"); envValue != "" {
		DB_HOST = envValue
	}
	if envValue := os.Getenv("DB_PORT"); envValue != "" {
		DB_PORT = envValue
	}
	if envValue := os.Getenv("DB_USER"); envValue != "" {
		DB_USER = envValue
	}
	if envValue := os.Getenv("DB_PASSWORD"); envValue != "" {
		DB_PASSWORD = envValue
	}
	if envValue := os.Getenv("DB_NAME"); envValue != "" {
		DB_NAME = envValue
	}
	if envValue := os.Getenv("DB_SSLMODE"); envValue != "" {
		DB_SSLMODE = envValue
	}
	if envValue := os.Getenv("REDIS_URL"); envValue != "" {
		REDIS_URL = envValue
	}

	if envValue := os.Getenv("JUDGE0_URL"); envValue != "" {
		JUDGE0_URL = envValue
	}
	if envValue := os.Getenv("JUDGE0_API_KEY"); envValue != "" {
		JUDGE0_API_KEY = envValue
	}
}
