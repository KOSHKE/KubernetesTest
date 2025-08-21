package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPageLimit extracts and validates page and limit query parameters
func GetPageLimit(c *gin.Context, defaultPage, defaultLimit, maxLimit int32) (int32, int32) {
	pageStr := c.DefaultQuery("page", strconv.FormatInt(int64(defaultPage), 10))
	limitStr := c.DefaultQuery("limit", strconv.FormatInt(int64(defaultLimit), 10))

	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil || page < 1 {
		page = int64(defaultPage)
	}

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit < 1 {
		limit = int64(defaultLimit)
	}
	if limit > int64(maxLimit) {
		limit = int64(maxLimit)
	}

	return int32(page), int32(limit)
}

// RequireUserID extracts and validates user_id from query parameters
// Returns error response if user_id is missing
func RequireUserID(c *gin.Context) (string, bool) {
	userID := c.Query("user_id")
	if userID == "" {
		RespondBadRequest(c, "user_id query parameter is required")
		return "", false
	}
	return userID, true
}

// ValidateRequest validates JSON request body and returns error response if invalid
func ValidateRequest(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		// Gin binding validation errors are more specific
		RespondBadRequest(c, "Validation failed: "+err.Error())
		return false
	}
	return true
}
