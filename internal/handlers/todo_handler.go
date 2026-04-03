package handlers

import (
	"net/http"
	"strconv"
	"todo-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateTodoInput struct {
	Title     string `json:"title" binding:"required"`
	Completed bool   `json:"completed"`
}

type UpdateTodoInput struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

// Helper: parse ID param
func parseIDParam(c *gin.Context) (int, bool) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return 0, false
	}
	return id, true
}

// CreateTodoHandler creates a new todo
func CreateTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdInterface, exist := c.Get("user_id")

		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		userID := userIdInterface.(string)

		var input CreateTodoInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		todo, err := repository.CreateTodo(pool, input.Title, input.Completed, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, todo)
	}
}

// GetAllTodosHandler returns all todos
func GetAllTodosHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIdInterface, exist := c.Get("user_id")

		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		userID := userIdInterface.(string)

		todos, err := repository.GetAllTodos(pool, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, todos)
	}
}

// GetTodoByIDHandler returns a todo by ID
func GetTodoByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIdInterface, exist := c.Get("user_id")

		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		userID := userIdInterface.(string)

		id, ok := parseIDParam(c)
		if !ok {
			return
		}

		todo, err := repository.GetTodoByID(pool, id, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, todo)
	}
}

// UpdateTodoByIDHandler updates a todo by ID
func UpdateTodoByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdInterface, exist := c.Get("user_id")

		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		userID := userIdInterface.(string)

		id, ok := parseIDParam(c)
		if !ok {
			return
		}

		var input UpdateTodoInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Title == nil && input.Completed == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field (title or completed) must be provided"})
			return
		}

		existTodo, err := repository.GetTodoByID(pool, id, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		title := existTodo.Title
		if input.Title != nil {
			title = *input.Title
		}

		completed := existTodo.Completed

		if input.Completed != nil {
			completed = *input.Completed
		}

		todo, err := repository.UpdateTodo(pool, id, title, completed, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, todo)
	}
}

func DeleteTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdInterface, exist := c.Get("user_id")

		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		userID := userIdInterface.(string)

		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid todo ID",
			})
			return // <--- stop execution here!
		}

		err = repository.DeleteToDo(pool, id, userID)
		if err != nil {
			if err.Error() == "todo with id "+idStr+" not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return // <--- stop execution here!
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Todo deleted successfully",
		})
	}
}
