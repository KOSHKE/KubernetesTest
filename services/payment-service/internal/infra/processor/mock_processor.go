package processor

import (
	"context"
	"time"

	"payment-service/internal/ports/processor"
)

// MockPaymentProcessor simulates payment outcomes based on card number
type MockPaymentProcessor struct{}

func NewMockPaymentProcessor() *MockPaymentProcessor { return &MockPaymentProcessor{} }

func (m *MockPaymentProcessor) Process(ctx context.Context, req processor.ProcessRequest) (processor.ProcessResult, error) {
	// Simulate processing latency with respect to context
	select {
	case <-time.After(300 * time.Millisecond):
	case <-ctx.Done():
		return processor.ProcessResult{Success: false, FailureReason: "context canceled"}, ctx.Err()
	}

	// Mock rules by card number
	switch req.CardNumber {
	case "", "0000000000000000":
		return processor.ProcessResult{Success: false, FailureReason: "missing card details"}, nil
	case "4000000000000002":
		return processor.ProcessResult{Success: false, FailureReason: "declined - insufficient funds"}, nil
	case "4000000000000119":
		return processor.ProcessResult{Success: false, FailureReason: "processing error"}, nil
	case "4000000000000341":
		return processor.ProcessResult{Success: false, FailureReason: "card expired"}, nil
	default:
		return processor.ProcessResult{Success: true}, nil
	}
}
