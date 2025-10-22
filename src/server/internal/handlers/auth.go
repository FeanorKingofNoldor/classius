package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/middleware"
	"github.com/classius/server/internal/models"
)

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Username   string `json:"username" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	FullName   string `json:"full_name,omitempty"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

// LoginRequest represents the user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
}

// UserResponse represents the user data returned in responses
type UserResponse struct {
	ID               uuid.UUID  `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	FullName         string     `json:"full_name"`
	AvatarURL        string     `json:"avatar_url"`
	SubscriptionTier string     `json:"subscription_tier"`
	LastActive       *time.Time `json:"last_active"`
	CreatedAt        time.Time  `json:"created_at"`
}

// ToUserResponse converts a User model to UserResponse
func ToUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		FullName:         user.FullName,
		AvatarURL:        user.AvatarURL,
		SubscriptionTier: user.SubscriptionTier,
		LastActive:       user.LastActive,
		CreatedAt:        user.CreatedAt,
	}
}

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "User already exists",
			"message": "A user with this email or username already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Password hashing failed",
			"message": "Unable to process password",
		})
		return
	}

	// Create user
	user := models.User{
		Username:         req.Username,
		Email:            req.Email,
		PasswordHash:     string(hashedPassword),
		FullName:         req.FullName,
		SubscriptionTier: "free",
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "User creation failed",
			"message": "Unable to create user account",
		})
		return
	}

	// Generate tokens
	accessToken, err := middleware.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Unable to generate access token",
		})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Unable to generate refresh token",
		})
		return
	}

	// Create user session (if device info provided)
	if req.DeviceID != "" {
		session := models.UserSession{
			UserID:       user.ID,
			DeviceID:     req.DeviceID,
			DeviceName:   req.DeviceName,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		db.DB.Create(&session)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"data": AuthResponse{
			User:         ToUserResponse(&user),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    24 * 60 * 60, // 24 hours in seconds
		},
	})
}

// Login handles user authentication
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Find user by email
	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid credentials",
				"message": "Email or password is incorrect",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"message": "Unable to process login request",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid credentials",
			"message": "Email or password is incorrect",
		})
		return
	}

	// Update last active time
	now := time.Now()
	user.LastActive = &now
	db.DB.Save(&user)

	// Generate tokens
	accessToken, err := middleware.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Unable to generate access token",
		})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Unable to generate refresh token",
		})
		return
	}

	// Create or update user session
	if req.DeviceID != "" {
		var session models.UserSession
		result := db.DB.Where("user_id = ? AND device_id = ?", user.ID, req.DeviceID).First(&session)
		
		if result.Error == gorm.ErrRecordNotFound {
			// Create new session
			session = models.UserSession{
				UserID:       user.ID,
				DeviceID:     req.DeviceID,
				DeviceName:   req.DeviceName,
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
			}
			db.DB.Create(&session)
		} else {
			// Update existing session
			session.AccessToken = accessToken
			session.RefreshToken = refreshToken
			session.ExpiresAt = time.Now().Add(24 * time.Hour)
			session.LastUsed = time.Now()
			if req.DeviceName != "" {
				session.DeviceName = req.DeviceName
			}
			db.DB.Save(&session)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data": AuthResponse{
			User:         ToUserResponse(&user),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    24 * 60 * 60, // 24 hours in seconds
		},
	})
}

// RefreshToken handles token refresh
func RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Validate refresh token
	claims, err := middleware.ValidateJWT(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid refresh token",
			"message": err.Error(),
		})
		return
	}

	// Get user from database
	var user models.User
	if err := db.DB.First(&user, "id = ?", claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not found",
			"message": "The user associated with this token no longer exists",
		})
		return
	}

	// Generate new tokens
	accessToken, err := middleware.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Unable to generate new access token",
		})
		return
	}

	newRefreshToken, err := middleware.GenerateRefreshToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Unable to generate new refresh token",
		})
		return
	}

	// Update user sessions with new tokens
	db.DB.Model(&models.UserSession{}).
		Where("user_id = ? AND refresh_token = ?", user.ID, req.RefreshToken).
		Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
			"expires_at":    time.Now().Add(24 * time.Hour),
			"last_used":     time.Now(),
		})

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
			"expires_in":    24 * 60 * 60,
		},
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	type LogoutRequest struct {
		DeviceID string `json:"device_id,omitempty"`
		AllDevices bool `json:"all_devices,omitempty"`
	}

	var req LogoutRequest
	c.ShouldBindJSON(&req) // Optional body

	// Get current user from middleware
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to perform this action",
		})
		return
	}

	if req.AllDevices {
		// Logout from all devices
		db.DB.Where("user_id = ?", user.ID).Delete(&models.UserSession{})
		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out from all devices successfully",
		})
	} else if req.DeviceID != "" {
		// Logout from specific device
		db.DB.Where("user_id = ? AND device_id = ?", user.ID, req.DeviceID).Delete(&models.UserSession{})
		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out from device successfully",
		})
	} else {
		// General logout (would need token blacklisting in production)
		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out successfully",
		})
	}
}