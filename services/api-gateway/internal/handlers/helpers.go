package handlers

import (
	"ecommerce-platform/services/api-gateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

// requireUserID extracts user ID from JWT context and returns error response if not found
func requireUserID(c *gin.Context) (string, bool) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{
			"error": gin.H{
				"code":    "unauthorized",
				"message": "User not authenticated",
			},
		})
		return "", false
	}
	return userID, true
}

// requireParam extracts URL parameter and returns error response if not found
func requireParam(c *gin.Context, paramName string) (string, bool) {
	paramValue := c.Param(paramName)
	if paramValue == "" {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "invalid_request",
				"message": paramName + " is required",
			},
		})
		return "", false
	}
	return paramValue, true
}

// validateJSONRequest validates JSON request body and returns error response if invalid
func validateJSONRequest(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "invalid_request",
				"message": "Invalid request format: " + err.Error(),
			},
		})
		return false
	}
	return true
}

// respondError sends structured error response
func respondError(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// respondSuccess sends structured success response
func respondSuccess(c *gin.Context, data interface{}, message string) {
	c.JSON(200, gin.H{
		"message": message,
		"data":    data,
	})
}

// respondCreated sends structured created response
func respondCreated(c *gin.Context, data interface{}, message string) {
	c.JSON(201, gin.H{
		"message": message,
		"data":    data,
	})
}
