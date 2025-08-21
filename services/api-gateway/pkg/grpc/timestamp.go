package grpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// FormatTimestamp formats protobuf timestamp to RFC3339 string
func FormatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}
