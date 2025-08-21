package http

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// ========== Domain Errors ==========

var (
	// User Domain Errors
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidUserData    = errors.New("invalid user data")
	ErrUserUpdateFailed   = errors.New("failed to update user")

	// Order Domain Errors
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderCreationFailed = errors.New("failed to create order")
	ErrOrderUpdateFailed   = errors.New("failed to update order")

	ErrInvalidOrderData  = errors.New("invalid order data")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrMissingUserID     = errors.New("user_id is required for order operations")

	// Payment Domain Errors
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrPaymentFailed       = errors.New("payment processing failed")
	ErrPaymentRefundFailed = errors.New("payment refund failed")
	ErrInvalidPaymentData  = errors.New("invalid payment data")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrCardDeclined        = errors.New("card declined")

	// Inventory Domain Errors
	ErrProductNotFound       = errors.New("product not found")
	ErrStockCheckFailed      = errors.New("stock check failed")
	ErrProductCreationFailed = errors.New("failed to create product")
	ErrProductUpdateFailed   = errors.New("failed to update product")
	ErrInvalidProductData    = errors.New("invalid product data")
	ErrStockUpdateFailed     = errors.New("failed to update stock")
)

// ========== Error Handlers ==========

// HandleUserClientError handles user client errors with appropriate HTTP status codes
func HandleUserClientError(c *gin.Context, err error, operation string) {
	if errors.Is(err, ErrUserNotFound) {
		RespondNotFound(c, "User not found")
		return
	}
	if errors.Is(err, ErrInvalidCredentials) {
		RespondUnauthorized(c, "Invalid credentials")
		return
	}
	if errors.Is(err, ErrEmailAlreadyExists) {
		RespondBadRequest(c, "Email already exists")
		return
	}
	if errors.Is(err, ErrInvalidUserData) {
		RespondBadRequest(c, "Invalid user data provided")
		return
	}
	if errors.Is(err, ErrUserUpdateFailed) {
		RespondInternalError(c, "Failed to update user profile")
		return
	}

	// Default case
	RespondInternalError(c, "Failed to "+operation)
}

// HandleOrderClientError handles order client errors with appropriate HTTP status codes
func HandleOrderClientError(c *gin.Context, err error, operation string) {
	if errors.Is(err, ErrOrderNotFound) {
		RespondNotFound(c, "Order not found")
		return
	}
	if errors.Is(err, ErrOrderCreationFailed) {
		RespondBadRequest(c, "Failed to create order")
		return
	}
	if errors.Is(err, ErrOrderUpdateFailed) {
		RespondBadRequest(c, "Failed to update order")
		return
	}

	if errors.Is(err, ErrInvalidOrderData) {
		RespondBadRequest(c, "Invalid order data provided")
		return
	}
	if errors.Is(err, ErrInsufficientStock) {
		RespondBadRequest(c, "Insufficient stock for requested items")
		return
	}
	if errors.Is(err, ErrMissingUserID) {
		RespondBadRequest(c, "User ID is required for order operations")
		return
	}

	// Default case
	RespondInternalError(c, "Failed to "+operation)
}

// HandlePaymentClientError handles payment client errors with appropriate HTTP status codes
func HandlePaymentClientError(c *gin.Context, err error, operation string) {
	if errors.Is(err, ErrPaymentNotFound) {
		RespondNotFound(c, "Payment not found")
		return
	}
	if errors.Is(err, ErrPaymentFailed) {
		RespondError(c, 402, "Payment processing failed")
		return
	}
	if errors.Is(err, ErrPaymentRefundFailed) {
		RespondBadRequest(c, "Failed to process refund")
		return
	}
	if errors.Is(err, ErrInvalidPaymentData) {
		RespondBadRequest(c, "Invalid payment data provided")
		return
	}
	if errors.Is(err, ErrInsufficientFunds) {
		RespondError(c, 402, "Insufficient funds")
		return
	}
	if errors.Is(err, ErrCardDeclined) {
		RespondError(c, 402, "Card declined")
		return
	}

	// Default case
	RespondInternalError(c, "Failed to "+operation)
}

// HandleInventoryClientError handles inventory client errors with appropriate HTTP status codes
func HandleInventoryClientError(c *gin.Context, err error, operation string) {
	if errors.Is(err, ErrProductNotFound) {
		RespondNotFound(c, "Product not found")
		return
	}
	if errors.Is(err, ErrStockCheckFailed) {
		RespondInternalError(c, "Failed to check stock availability")
		return
	}
	if errors.Is(err, ErrProductCreationFailed) {
		RespondBadRequest(c, "Failed to create product")
		return
	}
	if errors.Is(err, ErrProductUpdateFailed) {
		RespondBadRequest(c, "Failed to update product")
		return
	}
	if errors.Is(err, ErrInvalidProductData) {
		RespondBadRequest(c, "Invalid product data provided")
		return
	}
	if errors.Is(err, ErrStockUpdateFailed) {
		RespondInternalError(c, "Failed to update stock")
		return
	}

	// Default case
	RespondInternalError(c, "Failed to "+operation)
}

// HandleClientError maps client errors to appropriate HTTP status codes (generic fallback)
func HandleClientError(c *gin.Context, err error, notFoundMsg string) {
	// TODO: Add specific error types when they're defined in clients package
	// For now, use generic internal error
	RespondInternalError(c, "Service temporarily unavailable")
}
