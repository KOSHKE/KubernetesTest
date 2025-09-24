package handlers

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapGRPCErrorToHTTP maps gRPC error codes to HTTP status codes and error messages
func mapGRPCErrorToHTTP(err error) (int, string, string) {
	if err == nil {
		return 200, "success", ""
	}

	// Extract gRPC status
	st, ok := status.FromError(err)
	if !ok {
		// If it's not a gRPC status error, treat as internal server error
		return 500, "internal_error", "Internal server error"
	}

	// Map gRPC codes to HTTP status codes
	switch st.Code() {
	case codes.OK:
		return 200, "success", ""
	case codes.InvalidArgument:
		return 400, "invalid_argument", st.Message()
	case codes.Unauthenticated:
		return 401, "unauthenticated", st.Message()
	case codes.PermissionDenied:
		return 403, "permission_denied", st.Message()
	case codes.NotFound:
		return 404, "not_found", st.Message()
	case codes.AlreadyExists:
		return 409, "already_exists", st.Message()
	case codes.FailedPrecondition:
		return 412, "failed_precondition", st.Message()
	case codes.DeadlineExceeded:
		return 408, "request_timeout", st.Message()
	case codes.Unavailable:
		return 503, "service_unavailable", st.Message()
	case codes.Internal:
		return 500, "internal_error", st.Message()
	case codes.Unknown:
		return 500, "unknown_error", st.Message()
	default:
		return 500, "internal_error", "Internal server error"
	}
}

// handleGRPCError handles gRPC errors and sends appropriate HTTP response
func handleGRPCError(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}

	statusCode, code, message := mapGRPCErrorToHTTP(err)

	// For 4xx errors, use the gRPC message
	if statusCode >= 400 && statusCode < 500 {
		respondError(c, statusCode, code, message)
	} else {
		// For 5xx errors, use generic message for security
		respondError(c, statusCode, code, "Service temporarily unavailable")
	}

	return false
}
