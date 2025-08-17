package handlers

import (
	"net/http"
	"strconv"

	"api-gateway/internal/clients"
	"api-gateway/internal/types"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderClient     clients.OrderClient
	inventoryClient clients.InventoryClient
	paymentClient   clients.PaymentClient
}

func NewOrderHandler(orderClient clients.OrderClient, inventoryClient clients.InventoryClient, paymentClient clients.PaymentClient) *OrderHandler {
	return &OrderHandler{orderClient: orderClient, inventoryClient: inventoryClient, paymentClient: paymentClient}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = "demo-user"
	}
	var req struct {
		Items           []clients.OrderItemRequest `json:"items"`
		ShippingAddress string                     `json:"shipping_address"`
		PaymentDetails  clients.PaymentDetails     `json:"payment_details"`
		PaymentMethod   string                     `json:"payment_method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order must contain at least one item"})
		return
	}
	if req.ShippingAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shipping address is required"})
		return
	}
	stockReq := &clients.StockCheckRequest{Items: make([]clients.StockCheckItem, len(req.Items))}
	for i, item := range req.Items {
		stockReq.Items[i] = clients.StockCheckItem{ProductID: item.ProductID, Quantity: item.Quantity}
	}
	stockResponse, err := h.inventoryClient.CheckStock(c.Request.Context(), stockReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check stock availability"})
		return
	}
	if !stockResponse.AllAvailable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Some items are not available in requested quantities", "details": stockResponse.Results})
		return
	}
	orderReq := &clients.CreateOrderRequest{UserID: userID, Items: req.Items, ShippingAddress: req.ShippingAddress}
	order, err := h.orderClient.CreateOrder(c.Request.Context(), orderReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}
	paymentReq := &clients.ProcessPaymentRequest{OrderID: order.ID, UserID: userID, Amount: types.Money{Amount: order.TotalAmount.Amount, Currency: order.TotalAmount.Currency}, Method: req.PaymentMethod, Details: req.PaymentDetails}
	paymentResponse, err := h.paymentClient.ProcessPayment(c.Request.Context(), paymentReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}
	if !paymentResponse.Success {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Payment failed", "message": paymentResponse.Message})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"order": order, "payment": paymentResponse.Payment, "message": "Order created and payment processed successfully"})
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = "demo-user"
	}
	page := int32(1)
	limit := int32(10)
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = int32(l)
		}
	}
	orders, err := h.orderClient.GetUserOrders(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders, "page": page, "limit": limit, "total": len(orders)})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = "demo-user"
	}
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
	order, err := h.orderClient.GetOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = "demo-user"
	}
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
	order, err := h.orderClient.CancelOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": order, "message": "Order cancelled successfully"})
}
