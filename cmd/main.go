package main

import (
	"log"
	"todo-api/internal/config"
	"todo-api/internal/database"
	"todo-api/internal/handlers"
	"todo-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Defining an instance of configuration model with defining error var
	var cfg *config.Config
	var err error

	// Load configuration into cfg var
	cfg, err = config.Load()

	// Handling configuration errors
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Defining a connection pool to avoid repeated connections to DB
	var pool *pgxpool.Pool
	pool, err = database.Connect(cfg.DatabaseURL)

	// Handling DB connection error
	if err != nil {
		log.Fatal("Failed to load connect to database:", err)
	}

	// Defer pool closing to close at the end of main function
	defer pool.Close()

	// Defining router using Gin engine like borrowing a router functionality from Gin engine
	var router *gin.Engine = gin.Default()

	// Just to avoid tedious debugging logs
	router.SetTrustedProxies(nil)

	// Setting main route
	router.GET("/", func(c *gin.Context) {
		// gin.H => map[string]interface{}
		// gin.H => map[string]any{}
		c.JSON(200, gin.H{
			"message":  "Todo API is running",
			"status":   "success",
			"database": "connected",
		})
	})

	// Public routes
	router.POST("/auth/register", handlers.CreateUserHandler(pool))
	router.POST("/auth/login", handlers.LoginHandler(pool, cfg))

	// Defining a group for protected route and applying auth middleware
	protected := router.Group("/todos")
	protected.Use(middleware.AuthMiddleWare(cfg))

	// Protected routes
	{
		protected.POST("", handlers.CreateTodoHandler(pool))
		protected.GET("", handlers.GetAllTodosHandler(pool))
		protected.GET("/:id", handlers.GetTodoByIDHandler(pool))
		protected.PUT("/:id", handlers.UpdateTodoByIDHandler(pool))
		protected.DELETE("/:id", handlers.DeleteTodoHandler(pool))
	}

	// Middleware test route
	router.GET("/protected-test", middleware.AuthMiddleWare(cfg), handlers.TestProtectedHandler())

	// Running the server
	router.Run(":" + cfg.Port)
}
