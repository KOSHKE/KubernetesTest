package http

import (
	"errors"

	"github.com/gin-gonic/gin"
	commonErrors "ecommerce-platform/pkg/common/errors"
)

// ========== Error Handlers ==========

// HandleUserClientError handles user client errors with appropriate HTTP status codes
func HandleUserClientError(c *gin.Context, err error, operation string) {
	if errors.Is(err, commonErrors.ErrUserNotFound) {
		RespondNotFound(c, "User not found")
		return
	}
	if errors.Is(err, commonErrors.ErrInvalidCredentials) {
		RespondUnauthorized(c, "Invalid credentials")
		return
	}
	if errors.Is(err, commonErrors.ErrEmailAlreadyExists) {
		RespondBadRequest(c, "Email already exists")
		return
	}
	if errors.Is(err, commonErrors.ErrInvalidUserData) {
		RespondBadRequest(c, "Invalid user data provided")
		return
	}
	if errors.Is(err, commonErrors.ErrUserUpdateFailed) {
		RespondInternalError(c, "Failed to update user profile")
		return
	}

	// Default case
	RespondInternalError(c, "Failed to "+operation)
}

// HandleOrderClientError handles order client errors with appropriate HTTP status codes
func HandleOrderClientError(c *gin.Context, err error, operation string) {
	if errors.Is(err, commonErrors.ErrOrderNotFound) {
		RespondNotFound(c, "Order not found")
		return
	}
	if errors.Is(err, commonErrors.ErrOrderCreationFailed) {
		RespondBadRequest(c, "Failed to create order")
		return
	}
	if errors.Is(err, commonErrors.ErrOrderUpdateFailed) {
		RespondBadRequest(c, "Failed to update order")
		return
	}

	if errors.Is(err, commonErrors.ErrInvalidOrderData) {
		RespondBadRequest(c, "Invalid order data provided")
		return
	}
	if errors.Is(err, commonErrors.ErrInsufficientStock) {
		RespondBadRequest(c, "Insufficient stock for requested items")
		return
	}
	if errors.Is(err, commonErrors.ErrMissingUserID) {
		RespondBadRequest(c, "User ID is required for order operations")
		return
	}

	// Default case
	RespondInternalError(c, "Failed to "+operation)
}

// HandlePaymentClientError handles payment client errors with appropriate HTTP status codes
func HandlePaymentClientError(c *gin.Context, err error, operation string) {
	if errors.Is(err, commonErrors.ErrPaymentNotFound) {
		RespondNotFound(c, "Payment not found")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentValidationFailed) {
		RespondBadRequest(c, "Payment validation failed")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentCreationFailed) {
		RespondInternalError(c, "Failed to create payment")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentPersistenceFailed) {
		RespondInternalError(c, "Failed to persist payment")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentEventPublishFailed) {
		RespondInternalError(c, "Failed to publish payment event")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentEventMarshalFailed) {
		RespondInternalError(c, "Failed to marshal payment event")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentEventConsumerFailed) {
		RespondInternalError(c, "Failed to consume payment event")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentFailed) {
		RespondError(c, 402, "Payment processing failed")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentRefundFailed) {
		RespondError(c, 402, "Failed to process refund")
		return
	}
	if errors.Is(err, commonErrors.ErrInvalidPaymentData) {
		RespondBadRequest(c, "Invalid payment data provided")
		return
	}
	if errors.Is(err, commonErrors.ErrInsufficientFunds) {
		RespondError(c, 402, "Insufficient funds")
		return
	}
	if errors.Is(err, commonErrors.ErrCardDeclined) {
		RespondError(c, 402, "Card declined")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentDeclined) {
		RespondError(c, 402, "Payment declined")
		return
	}
	if errors.Is(err, commonErrors.ErrPaymentAlreadyProcessed) {
		RespondBadRequest(c, "Payment has already been processed")
		return
	}

	// Default case
	RespondInternalError(c, "Failed to "+operation)
}

// HandleInventoryClientError handles inventory client errors with appropriate HTTP status codes
func HandleInventoryClientError(c *gin.Context, err error, operation string) {
	if errors.Is(err, commonErrors.ErrProductNotFound) {
		RespondNotFound(c, "Product not found")
		return
	}
	if errors.Is(err, commonErrors.ErrStockCheckFailed) {
		RespondInternalError(c, "Failed to check stock availability")
		return
	}
	if errors.Is(err, commonErrors.ErrProductCreationFailed) {
		RespondBadRequest(c, "Failed to create product")
		return
	}
	if errors.Is(err, commonErrors.ErrProductUpdateFailed) {
		RespondBadRequest(c, "Failed to update product")
		return
	}
	if errors.Is(err, commonErrors.ErrInvalidProductData) {
		RespondBadRequest(c, "Invalid product data provided")
		return
	}
	if errors.Is(err, commonErrors.ErrStockUpdateFailed) {
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
