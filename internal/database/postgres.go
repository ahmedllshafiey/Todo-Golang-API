package database

import (
	"context" // Provides context.Context for managing deadlines, cancellations, and request-scoped values
	"log"     // Provides logging functions

	"github.com/jackc/pgx/v5/pgxpool" // PostgreSQL connection pool library
)

// Connect establishes a connection pool to a PostgreSQL database using the provided URL
func Connect(databaseURL string) (*pgxpool.Pool, error) {
	// Step 1: Initialize a base context
	// Context carries deadlines, cancellation signals, and other request-scoped values
	// It is passed to the connection pool to manage operations
	var ctx context.Context = context.Background()

	// Step 2: Prepare variables to hold connection configuration and errors
	// 'config' will hold the parsed configuration details from DATABASE_URL
	// 'err' will capture any errors during parsing or pool creation
	var config *pgxpool.Config // Initially nil
	var err error              // Initially nil

	// Step 3: Parse the database URL to generate a connection pool configuration
	// This step validates the URL and prepares all necessary parameters (host, port, user, password, etc.)
	config, err = pgxpool.ParseConfig(databaseURL)

	// Step 4: Handle errors that occur during parsing
	// For example, if DATABASE_URL is malformed or missing required fields
	if err != nil {
		log.Printf("Unable to parse DATABASE_URL: %v", err)
		return nil, err // Return nil pool and the error
	}

	// Step 5: Create a new connection pool using the parsed configuration and context
	// The pool manages multiple connections for efficient reuse and performance
	var pool *pgxpool.Pool
	pool, err = pgxpool.NewWithConfig(ctx, config)

	// Step 6: Handle errors that occur while creating the pool
	// This could be due to unreachable database, authentication failure, etc.
	if err != nil {
		log.Printf("Unable to create connection pool: %v", err)
		return nil, err // Return nil pool and the error
	}

	// Step 7: Verify that the pool can successfully connect to the database
	// Ping sends a simple query to ensure the database is reachable
	err = pool.Ping(ctx)

	// Step 8: Handle errors during ping
	// If ping fails, it indicates a problem with connectivity or credentials
	if err != nil {
		log.Printf("Unable to ping database: %v", err)
		pool.Close()    // Close the pool to clean up any partially opened connections
		return nil, err // Return nil pool and the error
	}

	// Step 9: Log successful connection
	// This confirms that the pool is ready for use in the application
	log.Printf("Successfully connected to PostgresSQL database")

	// Step 10: Return the initialized pool and nil error
	// The pool can now be used throughout the application for executing queries
	return pool, nil
}
