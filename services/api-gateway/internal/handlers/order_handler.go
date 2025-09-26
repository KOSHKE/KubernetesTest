package handlers

import (
	"context"
	"time"

	dto "ecommerce-platform/pkg/common/dto/order-service"
	"ecommerce-platform/services/api-gateway/internal/clients"

	"github.com/gin-gonic/gin"
)

const defaultRPCTimeout = 3 * time.Second

type OrderHandler struct {
	orderClient clients.OrderClient
}

func NewOrderHandler(orderClient clients.OrderClient) *OrderHandler {
	return &OrderHandler{
		orderClient: orderClient,
	}
}

// CreateOrder handles POST /api/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// Get user ID from JWT context
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	// Parse request
	var req dto.CreateOrderRequest
	if !validateJSONRequest(c, &req) {
		return
	}

	// Set user ID from JWT
	req.UserID = userID

	// Ensure bounded RPC time
	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultRPCTimeout)
	defer cancel()
	resp, err := h.orderClient.CreateOrder(ctx, &req)
	if !handleGRPCError(c, err) {
		return
	}

	respondCreated(c, gin.H{
		"order_id": resp.OrderID,
	}, resp.Message)
}

// GetOrder handles GET /api/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	// Get user ID from JWT context
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	// Get order ID from URL parameter
	orderID, ok := requireParam(c, "id")
	if !ok {
		return
	}

	// Ensure bounded RPC time
	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultRPCTimeout)
	defer cancel()
	resp, err := h.orderClient.GetOrder(ctx, orderID)
	if !handleGRPCError(c, err) {
		return
	}

	// Check if user owns this order
	if resp.UserID != userID {
		respondError(c, 403, "forbidden", "Access denied")
		return
	}

	respondSuccess(c, resp, "Order retrieved successfully")
}

// GetUserOrders handles GET /api/orders
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	// Get user ID from JWT context
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	// Ensure bounded RPC time
	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultRPCTimeout)
	defer cancel()
	resp, err := h.orderClient.GetUserOrders(ctx, userID)
	if !handleGRPCError(c, err) {
		return
	}

	respondSuccess(c, resp, "User orders retrieved successfully")
}
