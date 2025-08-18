package processor

import (
	"context"
	"math/rand"
	"time"

	"payment-service/internal/ports/processor"
)

// MockPaymentProcessor simulates payment outcomes (50% success) with latency
type MockPaymentProcessor struct{}

func NewMockPaymentProcessor() *MockPaymentProcessor { return &MockPaymentProcessor{} }

func (m *MockPaymentProcessor) Process(ctx context.Context, req processor.ProcessRequest) (processor.ProcessResult, error) {
	// Simulate processing latency with respect to context
	select {
	case <-time.After(300 * time.Millisecond):
	case <-ctx.Done():
		return processor.ProcessResult{Success: false, FailureReason: "context canceled"}, ctx.Err()
	}

	// 50% random outcome to mimic real-world uncertainty
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(2) == 0 {
		return processor.ProcessResult{Success: true}, nil
	}
	return processor.ProcessResult{Success: false, FailureReason: "payment declined (mock)"}, nil
}
