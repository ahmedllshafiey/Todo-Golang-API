package handlers

import (
	"errors"
	"net/http"
	"time"
	"todo-api/internal/config"
	"todo-api/internal/models"
	"todo-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Request structure for registering a new user
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`    // Required email field
	Password string `json:"password" binding:"required"` // Required password field
}

// Request structure for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`    // Required email field
	Password string `json:"password" binding:"required"` // Required password field
}

// Response structure for login containing JWT token
type LoginResponse struct {
	Token string `json:"token"` // JWT token string
}

// CreateUserHandler handles new user registration
func CreateUserHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind incoming JSON request to RegisterRequest struct
		var req RegisterRequest
		if err := c.BindJSON(&req); err != nil {
			// If JSON binding fails, respond with 400 Bad Request
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate password length (minimum 6 characters)
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters long"})
			return
		}

		// Hash the password using bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			// If hashing fails, return 500 Internal Server Error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Create a User model instance to store in the database
		user := models.User{
			Email:    req.Email,
			Password: string(hashedPassword), // Store hashed password
		}

		// Call repository function to insert new user into database
		createdUser, err := repository.CreateUser(pool, user)
		if err != nil {
			// Handle unique constraint violation (email already registered)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
				return
			}

			// Handle other database errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Respond with the created user and HTTP 201 Created
		c.JSON(http.StatusCreated, createdUser)
	}
}

// LoginHandler authenticates a user and returns a JWT token
func LoginHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind incoming JSON request to LoginRequest struct
		var loginRequest LoginRequest
		if err := c.BindJSON(&loginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Retrieve user from the database by email
		user, err := repository.GetUserByEmail(pool, loginRequest.Email)
		if err != nil {
			// If user not found, return 401 Unauthorized
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Compare stored hashed password with the provided password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
		if err != nil {
			// If password mismatch, return 401 Unauthorized
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Define JWT claims including user ID, email, and expiration
		claims := jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(), // Token valid for 24 hours
		}

		// Create a new JWT token with claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Sign the token using the secret from config
		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			// If signing fails, return 500 Internal Server Error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
			return
		}

		// Return the signed token in the response
		c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
	}
}

// TestProtectedHandler demonstrates a route protected by authentication middleware
func TestProtectedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve user ID from the request context (set by auth middleware)
		userID, exist := c.Get("user_id")
		if !exist {
			// If user ID missing, return 500 Internal Server Error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in the context"})
			return
		}

		// Respond with success message and user ID
		c.JSON(http.StatusOK, gin.H{
			"message": "Protected route accessed successfully",
			"user_id": userID,
		})
	}
}
