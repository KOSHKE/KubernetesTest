package grpcutils

import (
	stdErrors "errors"

	commonErrors "ecommerce-platform/pkg/common/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapErrorToStatus maps domain errors to gRPC status codes
func MapErrorToStatus(err error) error {
	switch {
	// User service errors
	case stdErrors.Is(err, commonErrors.ErrUserNotFound):
		return status.Errorf(codes.NotFound, "user not found")
	case stdErrors.Is(err, commonErrors.ErrInvalidCredentials):
		return status.Errorf(codes.Unauthenticated, "invalid credentials")
	case stdErrors.Is(err, commonErrors.ErrEmailAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "email already exists")
	case stdErrors.Is(err, commonErrors.ErrInvalidUserData):
		return status.Errorf(codes.InvalidArgument, "invalid user data")
	case stdErrors.Is(err, commonErrors.ErrUserCreationFailed):
		return status.Errorf(codes.Internal, "failed to create user")
	case stdErrors.Is(err, commonErrors.ErrTokenGenerationFailed):
		return status.Errorf(codes.Internal, "failed to generate tokens")
	case stdErrors.Is(err, commonErrors.ErrTokenStorageFailed):
		return status.Errorf(codes.Internal, "failed to store token")
	case stdErrors.Is(err, commonErrors.ErrTokenValidationFailed):
		return status.Errorf(codes.Unauthenticated, "token validation failed")
	case stdErrors.Is(err, commonErrors.ErrTokenExpired):
		return status.Errorf(codes.Unauthenticated, "token expired")
	case stdErrors.Is(err, commonErrors.ErrTokenRevoked):
		return status.Errorf(codes.Unauthenticated, "token revoked")
	case stdErrors.Is(err, commonErrors.ErrTokenInvalid):
		return status.Errorf(codes.Unauthenticated, "invalid token")
	case stdErrors.Is(err, commonErrors.ErrInvalidTokenClaims):
		return status.Errorf(codes.Unauthenticated, "invalid token claims")
	case stdErrors.Is(err, commonErrors.ErrTokenRevocationFailed):
		return status.Errorf(codes.Internal, "failed to revoke token")
	case stdErrors.Is(err, commonErrors.ErrRefreshTokenInvalid):
		return status.Errorf(codes.Unauthenticated, "invalid refresh token")
	case stdErrors.Is(err, commonErrors.ErrAuthenticationFailed):
		return status.Errorf(codes.Unauthenticated, "authentication failed")

	// Order service errors
	case stdErrors.Is(err, commonErrors.ErrOrderValidationFailed):
		return status.Errorf(codes.InvalidArgument, "order validation failed")
	case stdErrors.Is(err, commonErrors.ErrOrderNotFound):
		return status.Errorf(codes.NotFound, "order not found")
	case stdErrors.Is(err, commonErrors.ErrOrderAccessDenied):
		return status.Errorf(codes.PermissionDenied, "access denied")
	case stdErrors.Is(err, commonErrors.ErrOrderCreationFailed):
		return status.Errorf(codes.Internal, "failed to create order")
	case stdErrors.Is(err, commonErrors.ErrOrderUpdateFailed):
		return status.Errorf(codes.Internal, "failed to update order")
	case stdErrors.Is(err, commonErrors.ErrOrderPersistenceFailed):
		return status.Errorf(codes.Internal, "failed to persist order")
	case stdErrors.Is(err, commonErrors.ErrOrderStatusUpdateFailed):
		return status.Errorf(codes.Internal, "failed to update order status")
	case stdErrors.Is(err, commonErrors.ErrOrderCancellationFailed):
		return status.Errorf(codes.Internal, "failed to cancel order")
	case stdErrors.Is(err, commonErrors.ErrOrderRetrievalFailed):
		return status.Errorf(codes.Internal, "failed to retrieve order")
	case stdErrors.Is(err, commonErrors.ErrOrderItemNotFound):
		return status.Errorf(codes.NotFound, "order item not found")
	case stdErrors.Is(err, commonErrors.ErrOrderItemAdditionFailed):
		return status.Errorf(codes.Internal, "failed to add order item")
	case stdErrors.Is(err, commonErrors.ErrOrderItemRemovalFailed):
		return status.Errorf(codes.Internal, "failed to remove order item")
	case stdErrors.Is(err, commonErrors.ErrInvalidOrderData):
		return status.Errorf(codes.InvalidArgument, "invalid order data")
	case stdErrors.Is(err, commonErrors.ErrMissingUserID):
		return status.Errorf(codes.InvalidArgument, "user_id is required")
	case stdErrors.Is(err, commonErrors.ErrInvalidArgument):
		return status.Errorf(codes.InvalidArgument, "invalid argument")
	case stdErrors.Is(err, commonErrors.ErrInvalidOrderStatus):
		return status.Errorf(codes.InvalidArgument, "invalid order status")
	case stdErrors.Is(err, commonErrors.ErrInvalidOrderStatusTransition):
		return status.Errorf(codes.InvalidArgument, "invalid order status transition")
	case stdErrors.Is(err, commonErrors.ErrEmptyShippingAddress):
		return status.Errorf(codes.InvalidArgument, "shipping address cannot be empty")
	case stdErrors.Is(err, commonErrors.ErrInvalidQuantity):
		return status.Errorf(codes.InvalidArgument, "invalid quantity")
	case stdErrors.Is(err, commonErrors.ErrInvalidTransition):
		return status.Errorf(codes.InvalidArgument, "invalid transition")

	// Payment service errors
	case stdErrors.Is(err, commonErrors.ErrPaymentNotFound):
		return status.Errorf(codes.NotFound, "payment not found")
	case stdErrors.Is(err, commonErrors.ErrPaymentFailed):
		return status.Errorf(codes.Internal, "payment failed")
	case stdErrors.Is(err, commonErrors.ErrPaymentAlreadyProcessed):
		return status.Errorf(codes.FailedPrecondition, "payment already processed")
	case stdErrors.Is(err, commonErrors.ErrPaymentAmountInvalid):
		return status.Errorf(codes.InvalidArgument, "payment amount invalid")
	case stdErrors.Is(err, commonErrors.ErrPaymentMethodNotSupported):
		return status.Errorf(codes.InvalidArgument, "payment method not supported")

	// Inventory service errors
	case stdErrors.Is(err, commonErrors.ErrProductNotFound):
		return status.Errorf(codes.NotFound, "product not found")
	case stdErrors.Is(err, commonErrors.ErrInsufficientStock):
		return status.Errorf(codes.FailedPrecondition, "insufficient stock")
	case stdErrors.Is(err, commonErrors.ErrInventoryUpdateFailed):
		return status.Errorf(codes.Internal, "inventory update failed")
	case stdErrors.Is(err, commonErrors.ErrInventoryServiceUnavailable):
		return status.Errorf(codes.Unavailable, "inventory service unavailable")
	case stdErrors.Is(err, commonErrors.ErrInventoryConnectionFailed):
		return status.Errorf(codes.Unavailable, "failed to connect to inventory service")
	case stdErrors.Is(err, commonErrors.ErrInventoryRequestTimeout):
		return status.Errorf(codes.DeadlineExceeded, "inventory request timeout")
	case stdErrors.Is(err, commonErrors.ErrInventoryInvalidResponse):
		return status.Errorf(codes.Internal, "invalid response from inventory service")

	// Common value object errors
	case stdErrors.Is(err, commonErrors.ErrEmptyValue):
		return status.Errorf(codes.InvalidArgument, "value cannot be empty")
	case stdErrors.Is(err, commonErrors.ErrInvalidValue):
		return status.Errorf(codes.InvalidArgument, "invalid value")
	case stdErrors.Is(err, commonErrors.ErrValueTooLong):
		return status.Errorf(codes.InvalidArgument, "value too long")
	case stdErrors.Is(err, commonErrors.ErrValueTooShort):
		return status.Errorf(codes.InvalidArgument, "value too short")
	case stdErrors.Is(err, commonErrors.ErrEmptyCurrency):
		return status.Errorf(codes.InvalidArgument, "currency cannot be empty")
	case stdErrors.Is(err, commonErrors.ErrInvalidCurrencyLength):
		return status.Errorf(codes.InvalidArgument, "currency must be exactly 3 characters")
	case stdErrors.Is(err, commonErrors.ErrInvalidCurrencyFormat):
		return status.Errorf(codes.InvalidArgument, "currency must contain only letters")
	case stdErrors.Is(err, commonErrors.ErrCurrencyMismatch):
		return status.Errorf(codes.InvalidArgument, "currency mismatch")

	// Database and storage errors
	case stdErrors.Is(err, commonErrors.ErrDatabaseOperationFailed):
		return status.Errorf(codes.Internal, "database operation failed")
	case stdErrors.Is(err, commonErrors.ErrStorageURLRequired):
		return status.Errorf(codes.InvalidArgument, "storage URL is required")
	case stdErrors.Is(err, commonErrors.ErrStorageTimeoutInvalid):
		return status.Errorf(codes.InvalidArgument, "storage timeout must be positive")
	case stdErrors.Is(err, commonErrors.ErrInvalidStorageURL):
		return status.Errorf(codes.InvalidArgument, "invalid storage URL format")
	case stdErrors.Is(err, commonErrors.ErrStorageConnectionFailed):
		return status.Errorf(codes.Unavailable, "failed to connect to storage")
	case stdErrors.Is(err, commonErrors.ErrJWTManagerCreationFailed):
		return status.Errorf(codes.Internal, "failed to create JWT manager")

	default:
		return status.Errorf(codes.Internal, "internal server error: %v", err)
	}
}
