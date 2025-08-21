package handlers

import (
	"api-gateway/internal/clients"
	"api-gateway/pkg/http"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	http.BaseHandler
	paymentClient clients.PaymentClient
}

func NewPaymentHandler(paymentClient clients.PaymentClient) *PaymentHandler {
	return &PaymentHandler{paymentClient: paymentClient}
}

// ProcessPayment handles payment processing requests
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req http.ProcessPaymentRequest
	if !http.ValidateRequest(c, &req) {
		return
	}

	// Validate that user_id in request matches query parameter for security
	queryUserID := c.Query("user_id")
	if queryUserID == "" {
		http.RespondBadRequest(c, "user_id query parameter is required for payment operations")
		return
	}
	if queryUserID != req.UserID {
		http.RespondBadRequest(c, "user_id in request body must match user_id in query parameters")
		return
	}

	if h.HandlePaymentClientOperation(c, func() error {
		response, err := h.paymentClient.ProcessPayment(c.Request.Context(), req.ToClientRequest())
		if err != nil {
			return err
		}
		if !response.Success {
			http.RespondError(c, 402, "Payment failed")
			return nil // Not an error, just business logic
		}
		return nil
	}, "process payment") {
		http.RespondSuccess(c, gin.H{"message": "Payment processed successfully"}, "Payment processed successfully")
	}
}

// GetPayment retrieves payment information
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID, ok := h.RequireParam(c, "id")
	if !ok {
		return // Error response already sent by RequireParam
	}

	if h.HandlePaymentClientOperation(c, func() error {
		_, err := h.paymentClient.GetPayment(c.Request.Context(), paymentID)
		return err
	}, "get payment") {
		http.RespondSuccess(c, gin.H{"message": "Payment retrieved successfully"}, "Payment retrieved successfully")
	}
}

// RefundPayment processes payment refunds
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentID, ok := h.RequireParam(c, "id")
	if !ok {
		return // Error response already sent by RequireParam
	}

	var req http.RefundRequest
	if !h.ValidateRequest(c, &req) {
		return
	}

	if h.HandlePaymentClientOperation(c, func() error {
		_, err := h.paymentClient.RefundPayment(c.Request.Context(), paymentID, req.Amount, req.Reason)
		return err
	}, "process refund") {
		http.RespondSuccess(c, gin.H{"message": "Refund processed successfully"}, "Refund processed successfully")
	}
}

// GetPaymentMethods returns available payment methods (mock data)
func (h *PaymentHandler) GetPaymentMethods(c *gin.Context) {
	methods := []map[string]interface{}{
		{"id": "credit_card", "name": "Credit Card", "description": "Pay with Visa, MasterCard, or American Express", "enabled": true},
		{"id": "debit_card", "name": "Debit Card", "description": "Pay directly from your bank account", "enabled": true},
		{"id": "paypal", "name": "PayPal", "description": "Pay with your PayPal account", "enabled": false},
	}
	http.RespondSuccess(c, gin.H{"methods": methods}, "Mock payment methods - for demo purposes only")
}

// GetTestCards returns test card numbers for demo
func (h *PaymentHandler) GetTestCards(c *gin.Context) {
	testCards := []map[string]interface{}{
		{"number": "4111111111111111", "description": "Valid Visa card - payment succeeds", "type": "visa"},
		{"number": "5555555555554444", "description": "Valid MasterCard - payment succeeds", "type": "mastercard"},
		{"number": "4000000000000002", "description": "Declined card - insufficient funds", "type": "visa"},
		{"number": "4000000000000119", "description": "Processing error", "type": "visa"},
		{"number": "4000000000000341", "description": "Expired card", "type": "visa"},
	}
	http.RespondSuccess(c, gin.H{"test_cards": testCards}, "Test card numbers for demo - DO NOT USE IN PRODUCTION")
}
