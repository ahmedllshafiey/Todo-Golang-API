package handlers

import (
	"net/http"
	"strconv"
	"todo-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Input structure for creating a new todo item
type CreateTodoInput struct {
	Title     string `json:"title" binding:"required"` // Required title field
	Completed bool   `json:"completed"`                // Optional completed status
}

// Input structure for updating an existing todo item
type UpdateTodoInput struct {
	Title     *string `json:"title"`     // Optional pointer to title
	Completed *bool   `json:"completed"` // Optional pointer to completed
}

// Helper function to parse todo ID from URL parameter
func parseIDParam(c *gin.Context) (int, bool) {
	// Extract "id" parameter from request URL
	idStr := c.Param("id")

	// Attempt to convert string ID to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// If conversion fails, return a 400 Bad Request response
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return 0, false
	}

	// Return parsed ID and true for success
	return id, true
}

// CreateTodoHandler handles creation of a new todo
func CreateTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve user ID from request context (set by authentication middleware)
		userIdInterface, exist := c.Get("user_id")
		if !exist {
			// If user ID is missing, respond with 500 Internal Server Error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
			return
		}

		// Type assertion to convert user ID to string
		userID := userIdInterface.(string)

		// Initialize variable to hold JSON input
		var input CreateTodoInput

		// Bind incoming JSON to input structure, validating required fields
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Call repository function to create todo in database
		todo, err := repository.CreateTodo(pool, input.Title, input.Completed, userID)
		if err != nil {
			// If database operation fails, return 500 Internal Server Error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Respond with created todo and HTTP 201 Created status
		c.JSON(http.StatusCreated, todo)
	}
}

// GetAllTodosHandler returns all todos for the authenticated user
func GetAllTodosHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve user ID from request context
		userIdInterface, exist := c.Get("user_id")
		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
			return
		}

		userID := userIdInterface.(string)

		// Fetch all todos for the user from repository
		todos, err := repository.GetAllTodos(pool, userID)
		if err != nil {
			// Handle database errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the list of todos with HTTP 200 OK
		c.JSON(http.StatusOK, todos)
	}
}

// GetTodoByIDHandler returns a single todo by ID
func GetTodoByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve user ID from context
		userIdInterface, exist := c.Get("user_id")
		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
			return
		}
		userID := userIdInterface.(string)

		// Parse todo ID from URL parameter
		id, ok := parseIDParam(c)
		if !ok {
			return // Parsing failed; response already sent
		}

		// Fetch todo from repository by ID
		todo, err := repository.GetTodoByID(pool, id, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				// If no todo found, return 404 Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			// Handle other database errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the retrieved todo with HTTP 200 OK
		c.JSON(http.StatusOK, todo)
	}
}

// UpdateTodoByIDHandler updates an existing todo by ID
func UpdateTodoByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve user ID from context
		userIdInterface, exist := c.Get("user_id")
		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
			return
		}
		userID := userIdInterface.(string)

		// Parse todo ID from URL parameter
		id, ok := parseIDParam(c)
		if !ok {
			return
		}

		// Bind input JSON to update structure
		var input UpdateTodoInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure at least one field is provided to update
		if input.Title == nil && input.Completed == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field (title or completed) must be provided"})
			return
		}

		// Retrieve existing todo to preserve unchanged fields
		existTodo, err := repository.GetTodoByID(pool, id, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Use existing values if fields not provided in update
		title := existTodo.Title
		if input.Title != nil {
			title = *input.Title
		}
		completed := existTodo.Completed
		if input.Completed != nil {
			completed = *input.Completed
		}

		// Update todo in repository
		todo, err := repository.UpdateTodo(pool, id, title, completed, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return updated todo with HTTP 200 OK
		c.JSON(http.StatusOK, todo)
	}
}

// DeleteTodoHandler deletes a todo by ID
func DeleteTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve user ID from context
		userIdInterface, exist := c.Get("user_id")
		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
			return
		}
		userID := userIdInterface.(string)

		// Parse todo ID from URL parameter
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			// Handle invalid ID format
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
			return
		}

		// Delete todo in repository
		err = repository.DeleteToDo(pool, id, userID)
		if err != nil {
			if err.Error() == "todo with id "+idStr+" not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			// Handle other deletion errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Respond with success message
		c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
	}
}
