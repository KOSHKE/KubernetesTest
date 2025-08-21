package http

import (
	"github.com/gin-gonic/gin"
)

// BaseHandler provides common functionality for all HTTP handlers
type BaseHandler struct{}

// RequireUserID extracts and validates user_id from query parameters
// Returns (userID, success) - if !success, error response already sent
func (h *BaseHandler) RequireUserID(c *gin.Context) (string, bool) {
	return RequireUserID(c)
}

// RequireParam extracts and validates URL parameter
// Returns (paramValue, success) - if !success, error response already sent
func (h *BaseHandler) RequireParam(c *gin.Context, paramName string) (string, bool) {
	paramValue := c.Param(paramName)
	if paramValue == "" {
		RespondBadRequest(c, paramName+" is required")
		return "", false
	}
	return paramValue, true
}

// ValidateRequest validates JSON request body
// Returns success - if !success, error response already sent
func (h *BaseHandler) ValidateRequest(c *gin.Context, req interface{}) bool {
	return ValidateRequest(c, req)
}

// HandleClientOperation is a generic helper for client operations with error handling
func (h *BaseHandler) HandleClientOperation(c *gin.Context, operation func() error, errorHandler func(*gin.Context, error, string), operationName string) bool {
	if err := operation(); err != nil {
		errorHandler(c, err, operationName)
		return false
	}
	return true
}

// HandleUserClientOperation handles user client operations with proper error handling
func (h *BaseHandler) HandleUserClientOperation(c *gin.Context, operation func() error, operationName string) bool {
	return h.HandleClientOperation(c, operation, HandleUserClientError, operationName)
}

// HandleOrderClientOperation handles order client operations with proper error handling
func (h *BaseHandler) HandleOrderClientOperation(c *gin.Context, operation func() error, operationName string) bool {
	return h.HandleClientOperation(c, operation, HandleOrderClientError, operationName)
}

// HandlePaymentClientOperation handles payment client operations with proper error handling
func (h *BaseHandler) HandlePaymentClientOperation(c *gin.Context, operation func() error, operationName string) bool {
	return h.HandleClientOperation(c, operation, HandlePaymentClientError, operationName)
}

// HandleInventoryClientOperation handles inventory client operations with proper error handling
func (h *BaseHandler) HandleInventoryClientOperation(c *gin.Context, operation func() error, operationName string) bool {
	return h.HandleClientOperation(c, operation, HandleInventoryClientError, operationName)
}
