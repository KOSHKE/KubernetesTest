package handlers

import (
	"net/http"

	"api-gateway/internal/clients"
	"api-gateway/internal/types"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentClient clients.PaymentClient
}

func NewPaymentHandler(paymentClient clients.PaymentClient) *PaymentHandler {
	return &PaymentHandler{paymentClient: paymentClient}
}

// ProcessPayment handles payment processing requests
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req clients.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	// Basic validation
	if req.OrderID == "" || req.UserID == "" || req.Amount.Amount <= 0 || req.Amount.Currency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}
	if req.Details.CardNumber == "" || req.Details.CardHolder == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Card details required"})
		return
	}
	response, err := h.paymentClient.ProcessPayment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment processing failed"})
		return
	}
	if !response.Success {
		c.JSON(http.StatusPaymentRequired, response)
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetPayment retrieves payment information
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID required"})
		return
	}
	payment, err := h.paymentClient.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

// RefundPayment processes payment refunds
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID required"})
		return
	}
	var req struct {
		Amount types.Money `json:"amount" binding:"required"`
		Reason string      `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refund request"})
		return
	}
	if req.Amount.Amount <= 0 || req.Amount.Currency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}
	response, err := h.paymentClient.RefundPayment(c.Request.Context(), paymentID, req.Amount, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Refund processing failed"})
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetPaymentMethods returns available payment methods (mock data)
func (h *PaymentHandler) GetPaymentMethods(c *gin.Context) {
	methods := []map[string]interface{}{
		{"id": "credit_card", "name": "Credit Card", "description": "Pay with Visa, MasterCard, or American Express", "enabled": true},
		{"id": "debit_card", "name": "Debit Card", "description": "Pay directly from your bank account", "enabled": true},
		{"id": "paypal", "name": "PayPal", "description": "Pay with your PayPal account", "enabled": false},
	}
	c.JSON(http.StatusOK, gin.H{"methods": methods, "message": "Mock payment methods - for demo purposes only"})
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
	c.JSON(http.StatusOK, gin.H{"test_cards": testCards, "message": "Test card numbers for demo - DO NOT USE IN PRODUCTION", "note": "Use any future expiry date and any 3-digit CVV"})
}
