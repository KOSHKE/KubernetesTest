package processor

import "context"

// PaymentProcessor abstracts payment processing decision logic
type PaymentProcessor interface {
	Process(ctx context.Context, req ProcessRequest) (ProcessResult, error)
}

type ProcessRequest struct {
	CardNumber string
}

type ProcessResult struct {
	Success       bool
	FailureReason string
}
