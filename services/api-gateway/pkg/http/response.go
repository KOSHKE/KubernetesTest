package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standardized API response format
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RespondJSON sends a JSON response with the given status code and payload
func RespondJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

// RespondSuccess sends a successful JSON response
func RespondSuccess(c *gin.Context, data interface{}, message string) {
	RespondJSON(c, http.StatusOK, APIResponse{
		Data:    data,
		Message: message,
	})
}

// RespondCreated sends a created JSON response
func RespondCreated(c *gin.Context, data interface{}, message string) {
	RespondJSON(c, http.StatusCreated, APIResponse{
		Data:    data,
		Message: message,
	})
}

// RespondError sends an error JSON response
func RespondError(c *gin.Context, code int, message string) {
	RespondJSON(c, code, APIResponse{
		Error: message,
	})
}

// RespondBadRequest sends a bad request error response
func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, message)
}

// RespondUnauthorized sends an unauthorized error response
func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, message)
}

// RespondForbidden sends a forbidden error response
func RespondForbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, message)
}

// RespondNotFound sends a not found error response
func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, message)
}

// RespondInternalError sends an internal server error response
func RespondInternalError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, message)
}
