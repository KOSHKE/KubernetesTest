package grpcutils

import (
	"errors"
	"log"
	"strings"

	commonErrors "ecommerce-platform/pkg/common/errors"
	pkgvalidation "ecommerce-platform/pkg/validation"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errorMap contains predefined mappings of domain errors to gRPC codes and messages
var errorMap = map[error]struct {
	Code    codes.Code
	Message string
}{
	// User domain errors
	commonErrors.ErrUserNotFound:       {codes.NotFound, "user not found"},
	commonErrors.ErrInvalidCredentials: {codes.Unauthenticated, "invalid credentials"},
	commonErrors.ErrEmailAlreadyExists: {codes.AlreadyExists, "email already exists"},
	commonErrors.ErrInvalidUserData:    {codes.InvalidArgument, "invalid user data"},
	commonErrors.ErrUserCreationFailed: {codes.Internal, "failed to create user"},

	// Auth errors
	commonErrors.ErrTokenGenerationFailed: {codes.Internal, "failed to generate tokens"},
	commonErrors.ErrTokenStorageFailed:    {codes.Internal, "failed to store token"},
	commonErrors.ErrTokenValidationFailed: {codes.Unauthenticated, "token validation failed"},
	commonErrors.ErrTokenExpired:          {codes.Unauthenticated, "token expired"},
	commonErrors.ErrTokenRevoked:          {codes.Unauthenticated, "token revoked"},
	commonErrors.ErrTokenInvalid:          {codes.Unauthenticated, "invalid token"},
	commonErrors.ErrInvalidTokenClaims:    {codes.Unauthenticated, "invalid token claims"},
	commonErrors.ErrTokenRevocationFailed: {codes.Internal, "failed to revoke token"},
	commonErrors.ErrRefreshTokenInvalid:   {codes.Unauthenticated, "invalid refresh token"},
	commonErrors.ErrAuthenticationFailed:  {codes.Unauthenticated, "authentication failed"},

	// Configuration errors
	commonErrors.ErrAccessTokenSecretRequired:  {codes.InvalidArgument, "access token secret is required"},
	commonErrors.ErrRefreshTokenSecretRequired: {codes.InvalidArgument, "refresh token secret is required"},
	commonErrors.ErrAccessTokenTTLInvalid:      {codes.InvalidArgument, "access token TTL must be positive"},
	commonErrors.ErrRefreshTokenTTLInvalid:     {codes.InvalidArgument, "refresh token TTL must be positive"},
	commonErrors.ErrStorageURLRequired:         {codes.InvalidArgument, "storage URL is required"},
	commonErrors.ErrStorageTimeoutInvalid:      {codes.InvalidArgument, "storage timeout must be positive"},
	commonErrors.ErrInvalidStorageURL:          {codes.InvalidArgument, "invalid storage URL format"},
	commonErrors.ErrStorageConnectionFailed:    {codes.Unavailable, "failed to connect to storage"},
	commonErrors.ErrJWTManagerCreationFailed:   {codes.Internal, "failed to create JWT manager"},

	// Database errors
	commonErrors.ErrDatabaseOperationFailed: {codes.Internal, "database operation failed"},

	// Order domain errors
	commonErrors.ErrOrderNotFound:             {codes.NotFound, "order not found"},
	commonErrors.ErrOrderCreationFailed:       {codes.Internal, "failed to create order"},
	commonErrors.ErrOrderUpdateFailed:         {codes.Internal, "failed to update order"},
	commonErrors.ErrInvalidOrderData:          {codes.InvalidArgument, "invalid order data"},
	commonErrors.ErrMissingUserID:             {codes.InvalidArgument, "user_id is required for order operations"},
	commonErrors.ErrOrderAccessDenied:         {codes.PermissionDenied, "access denied"},
	commonErrors.ErrInvalidArgument:           {codes.InvalidArgument, "invalid argument"},
	commonErrors.ErrInvalidTransition:         {codes.InvalidArgument, "invalid transition"},
	commonErrors.ErrOrderItemProcessingFailed: {codes.Internal, "failed to process order items"},
	commonErrors.ErrOrderPersistenceFailed:    {codes.Internal, "failed to persist order"},
	commonErrors.ErrOrderFactoryFailed:        {codes.Internal, "failed to create order via factory"},
	commonErrors.ErrOrderValidationFailed:     {codes.InvalidArgument, "order validation failed"},
	commonErrors.ErrOrderItemNotFound:         {codes.NotFound, "order item not found"},
	commonErrors.ErrOrderItemRemovalFailed:    {codes.Internal, "failed to remove order item"},
	commonErrors.ErrOrderItemAdditionFailed:   {codes.Internal, "failed to add order item"},
	commonErrors.ErrOrderStatusUpdateFailed:   {codes.Internal, "failed to update order status"},
	commonErrors.ErrOrderCancellationFailed:   {codes.Internal, "failed to cancel order"},
	commonErrors.ErrOrderRetrievalFailed:      {codes.Internal, "failed to retrieve order"},
	commonErrors.ErrOrderEventPublishFailed:   {codes.Internal, "failed to publish order"},

	// Order value object domain errors
	commonErrors.ErrEmptyShippingAddress:         {codes.InvalidArgument, "shipping address cannot be empty"},
	commonErrors.ErrInvalidOrderStatus:           {codes.InvalidArgument, "invalid order status"},
	commonErrors.ErrInvalidOrderStatusTransition: {codes.InvalidArgument, "invalid order status transition"},

	// Payment domain errors
	commonErrors.ErrPaymentNotFound:           {codes.NotFound, "payment not found"},
	commonErrors.ErrPaymentFailed:             {codes.Internal, "payment failed"},
	commonErrors.ErrPaymentAlreadyProcessed:   {codes.FailedPrecondition, "payment already processed"},
	commonErrors.ErrPaymentAmountInvalid:      {codes.InvalidArgument, "payment amount invalid"},
	commonErrors.ErrPaymentMethodNotSupported: {codes.InvalidArgument, "payment method not supported"},
	commonErrors.ErrPaymentValidationFailed:   {codes.InvalidArgument, "payment validation failed"},
	commonErrors.ErrPaymentCreationFailed:     {codes.Internal, "failed to create payment"},
	commonErrors.ErrPaymentProcessingFailed:   {codes.Internal, "failed to process payment"},

	// Inventory domain errors
	commonErrors.ErrProductNotFound:           {codes.NotFound, "product not found"},
	commonErrors.ErrInvalidProductName:        {codes.InvalidArgument, "invalid product name"},
	commonErrors.ErrInvalidProductPrice:       {codes.InvalidArgument, "invalid product price"},
	commonErrors.ErrProductAlreadyExists:      {codes.AlreadyExists, "product already exists"},
	commonErrors.ErrProductCreationFailed:     {codes.Internal, "failed to create product"},
	commonErrors.ErrProductNotActive:          {codes.FailedPrecondition, "product is not active"},
	commonErrors.ErrProductNotAvailable:       {codes.FailedPrecondition, "product not available for purchase"},
	commonErrors.ErrInsufficientStock:         {codes.FailedPrecondition, "insufficient stock"},
	commonErrors.ErrInsufficientReservedStock: {codes.FailedPrecondition, "insufficient reserved stock"},
	commonErrors.ErrInvalidStock:              {codes.InvalidArgument, "invalid stock"},
	commonErrors.ErrStockNotFound:             {codes.NotFound, "stock not found"},
	commonErrors.ErrReservationNotFound:       {codes.NotFound, "reservation not found"},
	commonErrors.ErrReservationExpired:        {codes.FailedPrecondition, "reservation expired"},
	commonErrors.ErrInventoryUpdateFailed:     {codes.Internal, "inventory update failed"},

	// Category domain errors
	commonErrors.ErrCategoryNotFound:      {codes.NotFound, "category not found"},
	commonErrors.ErrInvalidCategoryID:     {codes.InvalidArgument, "invalid category ID"},
	commonErrors.ErrInvalidCategoryName:   {codes.InvalidArgument, "invalid category name"},
	commonErrors.ErrCategoryAlreadyExists: {codes.AlreadyExists, "category already exists"},
	commonErrors.ErrCategoryNotActive:     {codes.FailedPrecondition, "category is not active"},

	// Inventory client errors
	commonErrors.ErrInventoryServiceUnavailable: {codes.Unavailable, "inventory service unavailable"},
	commonErrors.ErrInventoryConnectionFailed:   {codes.Unavailable, "failed to connect to inventory service"},
	commonErrors.ErrInventoryRequestTimeout:     {codes.DeadlineExceeded, "inventory request timeout"},
	commonErrors.ErrInventoryInvalidResponse:    {codes.Internal, "invalid response from inventory service"},

	// Value object errors
	commonErrors.ErrEmptyValue:    {codes.InvalidArgument, "value cannot be empty"},
	commonErrors.ErrInvalidValue:  {codes.InvalidArgument, "invalid value"},
	commonErrors.ErrValueTooLong:  {codes.InvalidArgument, "value too long"},
	commonErrors.ErrValueTooShort: {codes.InvalidArgument, "value too short"},

	// Currency errors
	commonErrors.ErrEmptyCurrency:         {codes.InvalidArgument, "currency cannot be empty"},
	commonErrors.ErrInvalidCurrencyLength: {codes.InvalidArgument, "currency must be exactly 3 characters"},
	commonErrors.ErrInvalidCurrencyFormat: {codes.InvalidArgument, "currency must contain only letters"},
	commonErrors.ErrCurrencyMismatch:      {codes.InvalidArgument, "currency mismatch"},

	// General errors
	commonErrors.ErrInvalidRequest:     {codes.InvalidArgument, "invalid request"},
	commonErrors.ErrOperationFailed:    {codes.Internal, "operation failed"},
	commonErrors.ErrEventPublishFailed: {codes.Internal, "failed to publish event"},
	commonErrors.ErrInvalidProductID:   {codes.InvalidArgument, "invalid product ID"},
	commonErrors.ErrInvalidQuantity:    {codes.InvalidArgument, "invalid quantity"},
}

// MapErrorToStatus safely maps domain errors to gRPC status codes
func MapErrorToStatus(err error) error {
	if err == nil {
		return nil
	}

	// First, try to extract validation violations and attach as BadRequest details
	if violations, ok := pkgvalidation.ExtractViolations(err); ok {
		br := &errdetails.BadRequest{FieldViolations: make([]*errdetails.BadRequest_FieldViolation, 0, len(violations))}
		for _, v := range violations {
			br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       v.Field,
				Description: v.Description,
			})
		}

		st := status.New(codes.InvalidArgument, pkgvalidation.BuildMessage(violations))
		stWith, derr := st.WithDetails(br)
		if derr == nil {
			return stWith.Err()
		}
		return st.Err()
	}

	// Search in known error map
	for e, mapping := range errorMap {
		if errors.Is(err, e) {
			msg := mapping.Message
			if len(msg) > 0 {
				msg = strings.ToUpper(msg[:1]) + msg[1:]
			}
			return status.New(mapping.Code, msg).Err()
		}
	}

	// Log unknown errors for development
	log.Printf("[WARN] unmapped error: %v\n", err)

	// Return safe message to client
	return status.New(codes.Internal, "Internal server error").Err()
}
