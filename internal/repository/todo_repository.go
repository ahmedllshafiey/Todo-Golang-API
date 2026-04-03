package repository

import (
	"context" // Provides context for managing deadlines and cancellations
	"fmt"     // Provides formatting for error messages
	"time"    // Provides time.Duration for timeouts
	"todo-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool" // PostgreSQL connection pool
)

// CreateTodo inserts a new todo item into the database
func CreateTodo(pool *pgxpool.Pool, title string, completed bool, userID string) (*models.Todo, error) {
	// Step 1: Create a context with timeout for database operation
	// Ensures query fails if it takes longer than 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure context is canceled after operation

	// Step 2: Define SQL query to insert todo and return the inserted row
	query := `
		INSERT INTO todos (title, completed, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, completed, created_at, updated_at, user_id
	`

	// Step 3: Define a variable to hold the returned todo
	var todo models.Todo

	// Step 4: Execute query with provided parameters and scan returned row into todo
	err := pool.QueryRow(ctx, query, title, completed, userID).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)

	// Step 5: Handle any query errors
	if err != nil {
		return nil, err
	}

	// Step 6: Return the created todo and nil error
	return &todo, nil
}

// GetAllTodos retrieves all todos for a specific user
func GetAllTodos(pool *pgxpool.Pool, userID string) ([]models.Todo, error) {
	// Step 1: Create a context with 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 2: SQL query to select all todos for the user, ordered by creation date
	query := `
		SELECT id, title, completed, created_at, updated_at, user_id
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	// Step 3: Execute query and get rows
	rows, err := pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure rows are closed after iteration

	// Step 4: Initialize a slice to store todos
	var todos []models.Todo

	// Step 5: Iterate through rows and scan each into a Todo struct
	for rows.Next() {
		var todo models.Todo

		err = rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&todo.UserID,
		)
		if err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	// Step 6: Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Step 7: Return the slice of todos
	return todos, nil
}

// GetTodoByID retrieves a single todo by ID for a specific user
func GetTodoByID(pool *pgxpool.Pool, id int, userID string) (*models.Todo, error) {
	// Step 1: Create context with 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 2: SQL query to select a specific todo by ID and user ID
	query := `
		SELECT id, title, completed, created_at, updated_at, user_id
		FROM todos
		WHERE id = $1 AND user_id = $2
	`

	// Step 3: Define a variable to store the retrieved todo
	var todo models.Todo

	// Step 4: Execute query and scan result into todo
	err := pool.QueryRow(ctx, query, id, userID).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)

	// Step 5: Handle errors (e.g., todo not found)
	if err != nil {
		return nil, err
	}

	// Step 6: Return the retrieved todo
	return &todo, nil
}

// UpdateTodo updates an existing todo for a user
func UpdateTodo(pool *pgxpool.Pool, id int, title string, completed bool, userID string) (*models.Todo, error) {
	// Step 1: Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 2: SQL query to update a todo and return the updated row
	query := `
		UPDATE todos
		SET title = $1, completed = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND user_id = $4
		RETURNING id, title, completed, created_at, updated_at, user_id
	`

	// Step 3: Variable to store the updated todo
	var todo models.Todo

	// Step 4: Execute update query and scan returned row into todo
	err := pool.QueryRow(ctx, query, title, completed, id, userID).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)
	if err != nil {
		return nil, err
	}

	// Step 5: Return updated todo
	return &todo, nil
}

// DeleteToDo removes a todo by ID for a specific user
func DeleteToDo(pool *pgxpool.Pool, id int, userID string) error {
	// Step 1: Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 2: SQL query to delete todo by ID and user ID
	query := `
		DELETE FROM todos
		WHERE id = $1 AND user_id = $2
	`

	// Step 3: Execute delete query
	commandTag, err := pool.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	// Step 4: Check if any row was affected (i.e., todo existed)
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("Todo with id %d not found", id)
	}

	// Step 5: Return nil if deletion succeeded
	return nil
}
