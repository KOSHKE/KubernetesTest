package handlers

import (
	"ecommerce-platform/services/api-gateway/internal/clients"
	"ecommerce-platform/services/api-gateway/internal/middleware"
	"ecommerce-platform/services/api-gateway/pkg/http"

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
	userID, ok := middleware.GetUserID(c)
	if !ok {
		http.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req http.ProcessPaymentRequest
	if !http.ValidateRequest(c, &req) {
		return
	}

	// Set user ID from JWT context
	req.UserID = userID

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
