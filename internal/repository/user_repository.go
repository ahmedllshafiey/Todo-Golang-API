package repository

import (
	"context" // Provides context for managing timeouts and cancellations
	"time"    // Provides time.Duration for timeout values
	"todo-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool" // PostgreSQL connection pool
)

// CreateUser inserts a new user into the database
func CreateUser(pool *pgxpool.Pool, user models.User) (*models.User, error) {
	// Step 1: Create a context with a 5-second timeout
	// Ensures the database operation will not hang indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Cancel the context when function completes

	// Step 2: Define SQL query to insert a new user
	// RETURNING clause retrieves the inserted row fields
	query := `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
		RETURNING id, email, created_at, updated_at
	`

	// Step 3: Execute the query with provided email and password
	// QueryRow returns a single row result
	err := pool.QueryRow(ctx, query, user.Email, user.Password).Scan(
		&user.ID,        // Scan the returned ID into user.ID
		&user.Email,     // Scan returned email (confirmation)
		&user.CreatedAt, // Scan created_at timestamp
		&user.UpdatedAt, // Scan updated_at timestamp
	)

	// Step 4: Handle any error from insertion
	// Could be a constraint violation (unique email) or connection issue
	if err != nil {
		return nil, err
	}

	// Step 5: Return the newly created user and nil error
	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func GetUserByEmail(pool *pgxpool.Pool, email string) (*models.User, error) {
	// Step 1: Create context with 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 2: Define SQL query to select user by email
	query := `
		SELECT id, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	// Step 3: Prepare a variable to hold the retrieved user
	var user models.User

	// Step 4: Execute the query and scan the result into user struct
	err := pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Step 5: Handle any error (e.g., no user found)
	if err != nil {
		return nil, err
	}

	// Step 6: Return the user and nil error
	return &user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(pool *pgxpool.Pool, id string) (*models.User, error) {
	// Step 1: Create context with 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 2: Define SQL query to select user by ID
	query := `
		SELECT id, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	// Step 3: Prepare a variable to hold the retrieved user
	var user models.User

	// Step 4: Execute the query and scan the result into user struct
	err := pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Step 5: Handle any error (e.g., no user found)
	if err != nil {
		return nil, err
	}

	// Step 6: Return the user and nil error
	return &user, nil
}
