package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResponseEnvelope represents the standard API response structure
type ResponseEnvelope struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
	Meta   interface{} `json:"meta,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ResponseEnvelope{
		Status: "success",
		Data:   data,
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, ResponseEnvelope{
		Status: "success",
		Data:   data,
		Meta:   meta,
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, ResponseEnvelope{
		Status: "success",
		Data:   data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ResponseEnvelope{
		Status: "error",
		Error:  message,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "Bad request"
	}
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	Error(c, http.StatusForbidden, message)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(c, http.StatusNotFound, message)
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(c, http.StatusInternalServerError, message)
}

// ValidationError sends a 422 Unprocessable Entity response
func ValidationError(c *gin.Context, message string) {
	if message == "" {
		message = "Validation error"
	}
	Error(c, http.StatusUnprocessableEntity, message)
}

// ValidationErrorDetails sends a 422 response with validation details
func ValidationErrorDetails(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusUnprocessableEntity, ResponseEnvelope{
		Status: "error",
		Error:  "Validation failed",
		Data:   errors,
	})
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	Error(c, http.StatusServiceUnavailable, message)
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "Too many requests"
	}
	Error(c, http.StatusTooManyRequests, message)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	Error(c, http.StatusConflict, message)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data interface{}, page, pageSize, total int64) {
	totalPages := int64(1)
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	c.JSON(http.StatusOK, ResponseEnvelope{
		Status: "success",
		Data:   data,
		Meta: gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}
