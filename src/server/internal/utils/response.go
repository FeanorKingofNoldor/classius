package utils

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse sends a successful response
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, response)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	response := APIResponse{
		Success: false,
		Message: message,
	}

	// Include error details in development/debug mode
	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(statusCode, response)
}

// GetIntQuery parses an integer query parameter with default value and bounds
func GetIntQuery(c *gin.Context, key string, defaultValue, min, max int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	if value < min {
		return min
	}
	if value > max {
		return max
	}

	return value
}

// GetBoolQuery parses a boolean query parameter with default value
func GetBoolQuery(c *gin.Context, key string, defaultValue bool) bool {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse wraps data with pagination info
type PaginatedResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginatedSuccessResponse sends a paginated successful response
func PaginatedSuccessResponse(c *gin.Context, message string, data interface{}, pagination PaginationResponse) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data: PaginatedResponse{
			Data:       data,
			Pagination: pagination,
		},
	}
	c.JSON(http.StatusOK, response)
}