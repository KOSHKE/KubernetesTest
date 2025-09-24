package middleware

import (
	"strings"

	"ecommerce-platform/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT access token
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": "Authorization header required",
				},
			})
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		// Set user info and full claims in context
		c.Set("user_id", claims.UserID)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	if id, ok := userID.(string); ok {
		return id, true
	}

	return "", false
}

// GetClaims extracts full JWT claims from context
func GetClaims(c *gin.Context) (*jwt.Claims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	if jwtClaims, ok := claims.(*jwt.Claims); ok {
		return jwtClaims, true
	}

	return nil, false
}
