package errors

import "errors"

// Common errors used across all services
var (
	// User domain errors
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidUserData    = errors.New("invalid user data")
	ErrUserCreationFailed = errors.New("failed to create user")

	// Auth errors
	ErrTokenGenerationFailed = errors.New("token generation failed")
	ErrTokenStorageFailed    = errors.New("token storage operation failed")
	ErrTokenValidationFailed = errors.New("token validation failed")
	ErrTokenExpired          = errors.New("token expired")
	ErrTokenRevoked          = errors.New("token revoked")
	ErrTokenInvalid          = errors.New("invalid token")
	ErrInvalidTokenClaims    = errors.New("invalid token claims")
	ErrTokenRevocationFailed = errors.New("failed to revoke token")
	ErrRefreshTokenInvalid   = errors.New("invalid refresh token")
	ErrAuthenticationFailed  = errors.New("authentication failed")

	// Configuration errors
	ErrAccessTokenSecretRequired  = errors.New("access token secret is required")
	ErrRefreshTokenSecretRequired = errors.New("refresh token secret is required")
	ErrAccessTokenTTLInvalid      = errors.New("access token TTL must be positive")
	ErrRefreshTokenTTLInvalid     = errors.New("refresh token TTL must be positive")
	ErrStorageURLRequired         = errors.New("storage URL is required")
	ErrStorageTimeoutInvalid      = errors.New("storage timeout must be positive")
	ErrInvalidStorageURL          = errors.New("invalid storage URL format")
	ErrStorageConnectionFailed    = errors.New("failed to connect to storage")
	ErrJWTManagerCreationFailed   = errors.New("failed to create JWT manager")

	// Database errors
	ErrDatabaseOperationFailed = errors.New("database operation failed")

	// Order domain errors
	ErrOrderNotFound             = errors.New("order not found")
	ErrOrderCreationFailed       = errors.New("failed to create order")
	ErrOrderUpdateFailed         = errors.New("failed to update order")
	ErrInvalidOrderData          = errors.New("invalid order data")
	ErrMissingUserID             = errors.New("user_id is required for order operations")
	ErrOrderAccessDenied         = errors.New("access denied")
	ErrInvalidArgument           = errors.New("invalid argument")
	ErrInvalidTransition         = errors.New("invalid status transition")
	ErrOrderItemProcessingFailed = errors.New("failed to process order items")
	ErrOrderPersistenceFailed    = errors.New("failed to persist order")
	ErrOrderFactoryFailed        = errors.New("failed to create order via factory")
	ErrOrderValidationFailed     = errors.New("order validation failed")
	ErrOrderItemNotFound         = errors.New("order item not found")
	ErrOrderItemRemovalFailed    = errors.New("failed to remove order item")
	ErrOrderItemAdditionFailed   = errors.New("failed to add order item")
	ErrOrderStatusUpdateFailed   = errors.New("failed to update order status")
	ErrOrderCancellationFailed   = errors.New("failed to cancel order")
	ErrOrderRetrievalFailed      = errors.New("failed to retrieve order")
	ErrOrderEventPublishFailed   = errors.New("failed to publish order event")

	// Order value object domain errors
	ErrEmptyShippingAddress         = errors.New("shipping address cannot be empty")
	ErrInvalidOrderStatus           = errors.New("invalid order status")
	ErrInvalidOrderStatusTransition = errors.New("invalid order status transition")

	// Payment domain errors
	ErrPaymentNotFound           = errors.New("payment not found")
	ErrPaymentFailed             = errors.New("payment failed")
	ErrPaymentAlreadyProcessed   = errors.New("payment already processed")
	ErrPaymentAmountInvalid      = errors.New("payment amount invalid")
	ErrPaymentMethodNotSupported = errors.New("payment method not supported")
	ErrPaymentValidationFailed   = errors.New("payment validation failed")
	ErrPaymentCreationFailed     = errors.New("failed to create payment")
	ErrPaymentProcessingFailed   = errors.New("failed to process payment")

	// Inventory domain errors
	ErrProductNotFound           = errors.New("product not found")
	ErrInvalidProductName        = errors.New("invalid product name")
	ErrInvalidProductPrice       = errors.New("invalid product price")
	ErrProductAlreadyExists      = errors.New("product already exists")
	ErrProductNotActive          = errors.New("product is not active")
	ErrProductNotAvailable       = errors.New("product not available for purchase")
	ErrInsufficientStock         = errors.New("insufficient stock")
	ErrInsufficientReservedStock = errors.New("insufficient reserved stock")
	ErrInvalidStock              = errors.New("invalid stock")
	ErrStockNotFound             = errors.New("stock not found")
	ErrReservationNotFound       = errors.New("reservation not found")
	ErrReservationExpired        = errors.New("reservation expired")
	ErrInventoryUpdateFailed     = errors.New("inventory update failed")

	// Category domain errors
	ErrCategoryNotFound      = errors.New("category not found")
	ErrInvalidCategoryID     = errors.New("invalid category ID")
	ErrInvalidCategoryName   = errors.New("invalid category name")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryNotActive     = errors.New("category is not active")

	// Inventory client errors
	ErrInventoryServiceUnavailable = errors.New("inventory service unavailable")
	ErrInventoryConnectionFailed   = errors.New("failed to connect to inventory service")
	ErrInventoryRequestTimeout     = errors.New("inventory request timeout")
	ErrInventoryInvalidResponse    = errors.New("invalid response from inventory service")

	// Value object errors
	ErrEmptyValue    = errors.New("value cannot be empty")
	ErrInvalidValue  = errors.New("invalid value")
	ErrValueTooLong  = errors.New("value too long")
	ErrValueTooShort = errors.New("value too short")

	// Currency errors
	ErrEmptyCurrency         = errors.New("currency cannot be empty")
	ErrInvalidCurrencyLength = errors.New("currency must be exactly 3 characters")
	ErrInvalidCurrencyFormat = errors.New("currency must contain only letters")
	ErrCurrencyMismatch      = errors.New("currency mismatch")

	// General errors
	ErrInvalidRequest     = errors.New("invalid request")
	ErrOperationFailed    = errors.New("operation failed")
	ErrEventPublishFailed = errors.New("failed to publish event")
	ErrInvalidProductID   = errors.New("invalid product ID")
	ErrInvalidQuantity    = errors.New("invalid quantity")
)
