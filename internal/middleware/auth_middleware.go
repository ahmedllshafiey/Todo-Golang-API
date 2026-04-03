package middleware

import (
	"fmt"      // For formatting error messages
	"net/http" // Provides HTTP status codes
	"strings"  // String utilities (TrimPrefix)
	"time"     // Time utilities for expiration handling
	"todo-api/internal/config"

	"github.com/gin-gonic/gin"  // Gin web framework
	"github.com/golang-jwt/jwt" // JWT handling
)

// AuthMiddleWare validates JWT tokens and ensures requests are authenticated
func AuthMiddleWare(cfg *config.Config) gin.HandlerFunc {
	// Step 1: Return a gin.HandlerFunc (closure) so we can inject config
	return func(c *gin.Context) {
		// Step 2: Retrieve the Authorization header from the request
		authHeader := c.GetHeader("Authorization")

		// Step 3: Check if Authorization header is present
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort() // Stop processing further handlers
			return
		}

		// Step 4: Remove "Bearer " prefix to get the raw token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Step 5: Ensure the token was actually provided after "Bearer "
		if tokenString == "" || tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Step 6: Parse the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Step 6a: Ensure token uses HS256 signing method
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// Step 6b: Return the secret key for validation
			return []byte(cfg.JWTSecret), nil
		})

		// Step 7: Handle parsing errors or invalid token
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Step 8: Extract claims from the token (payload)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token payload"})
			c.Abort()
			return
		}

		// Step 9: Retrieve user_id from claims
		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token payload"})
			c.Abort()
			return
		}

		// Step 10: Check expiration (exp claim) if present
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			if time.Now().After(expirationTime) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
				c.Abort()
				return
			}
		}

		// Step 11: Store user_id in the Gin context for downstream handlers
		c.Set("user_id", userID)

		// Step 12: Continue to the next middleware/handler
		c.Next()
	}
}
