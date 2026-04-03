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
	var cfg *config.Config
	var err error

	cfg, err = config.Load()

	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	var pool *pgxpool.Pool
	pool, err = database.Connect(cfg.DatabaseURL)

	if err != nil {
		log.Fatal("Failed to load connect to database:", err)
	}

	defer pool.Close()

	var router *gin.Engine = gin.Default()
	router.SetTrustedProxies(nil)
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

	router.Run(":" + cfg.Port)
}
