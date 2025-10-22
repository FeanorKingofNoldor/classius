package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"

	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/models"
)
// CORS middleware with appropriate settings for development and production
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()

	// Configure based on environment
	if viper.GetString("environment") == "production" {
		config.AllowOrigins = []string{
			"https://classius.com",
			"https://app.classius.com",
		}
	} else {
		config.AllowAllOrigins = true
	}

	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"Accept",
		"Cache-Control",
	}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
}

// JWTClaims represents the claims in our JWT tokens
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// GetJWTSecret returns the JWT secret from configuration
func GetJWTSecret() []byte {
	secret := viper.GetString("jwt.secret")
	if secret == "" {
		secret = "classius-dev-secret-key-change-in-production"
	}
	return []byte(secret)
}

// GenerateJWT generates a JWT token for the user
func GenerateJWT(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "classius-server",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecret())
}

// GenerateRefreshToken generates a refresh token for the user
func GenerateRefreshToken(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "classius-server",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecret())
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return GetJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AuthRequired middleware that validates JWT tokens
func AuthRequired() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "Please provide a valid authentication token",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization header format",
				"message": "Authorization header must use Bearer token format",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Get user from database
		var user models.User
		if err := db.DB.First(&user, "id = ?", claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not found",
				"message": "The user associated with this token no longer exists",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user", &user)
		c.Set("user_id", user.ID.String())
		c.Set("username", user.Username)

		c.Next()
	})
}

// GetCurrentUser extracts the current user from the Gin context
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u, true
		}
	}
	return nil, false
}

// RateLimiting middleware (basic implementation)
func RateLimiting() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Simple rate limiting using Redis (if available)
		if db.Redis == nil {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := "rate_limit:" + clientIP

		// Check current count
		val, err := db.Redis.Get(c.Request.Context(), key).Result()
		if err != nil && err.Error() != "redis: nil" {
			// Redis error, allow request
			c.Next()
			return
		}

		// Parse current count
		var count int64 = 0
		if val != "" {
			count = 1 // Simplified - in production, parse the actual count
		}

		// Check if rate limit exceeded
		limit := int64(100) // 100 requests per minute
		if count >= limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		// Increment counter
		db.Redis.Incr(c.Request.Context(), key)
		db.Redis.Expire(c.Request.Context(), key, time.Minute)

		c.Next()
	})
}

// RequestLogger middleware for detailed request logging
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[%s] %s %s %d %s %s %s\n",
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.ClientIP,
				param.Method,
				param.StatusCode,
				param.Path,
				param.Latency,
				param.ErrorMessage,
			)
		},
		Output:    gin.DefaultWriter,
		SkipPaths: []string{"/health"},
	})
}

// Security headers middleware
func SecurityHeaders() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		if viper.GetString("environment") == "production" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		c.Next()
	})
}