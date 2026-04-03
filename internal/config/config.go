package config

import (
	"log" // Provides logging functionality
	"os"  // Provides access to environment variables

	"github.com/joho/godotenv" // Library to load .env file into environment
)

// Config defines the application configuration structure
// It holds essential environment variables used across the application
type Config struct {
	DatabaseURL string // PostgreSQL connection URL
	Port        string // HTTP server port
	JWTSecret   string // Secret key used to sign JWT tokens
}

// Load reads environment variables and returns a Config instance
func Load() (*Config, error) {
	// Step 1: Load environment variables from a .env file if it exists
	// godotenv.Load() reads the file and populates os.Getenv automatically
	// It returns an error if the file is missing or unreadable
	var err error = godotenv.Load()

	// Step 2: Handle any errors from loading .env
	// If the .env file does not exist, we continue using environment variables
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Step 3: Create a new Config struct and populate it from environment variables
	// os.Getenv returns an empty string if the variable is not set
	var config *Config = &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"), // Fetch database URL
		Port:        os.Getenv("PORT"),         // Fetch server port
		JWTSecret:   os.Getenv("JWT_SECRET"),   // Fetch JWT signing secret
	}

	// Step 4: Return the populated configuration and a nil error
	// The returned config can now be used throughout the application
	return config, nil
}
