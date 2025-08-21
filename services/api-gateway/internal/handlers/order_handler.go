package handlers

import (
	"api-gateway/internal/clients"
	"api-gateway/pkg/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	http.BaseHandler
	orderClient     clients.OrderClient
	inventoryClient clients.InventoryClient
	paymentClient   clients.PaymentClient
}

func NewOrderHandler(orderClient clients.OrderClient, inventoryClient clients.InventoryClient, paymentClient clients.PaymentClient) *OrderHandler {
	return &OrderHandler{orderClient: orderClient, inventoryClient: inventoryClient, paymentClient: paymentClient}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req http.CreateOrderRequest
	if !http.ValidateRequest(c, &req) {
		return
	}

	// Check stock availability
	var stockResponse *clients.StockCheckResponse
	if !h.HandleInventoryClientOperation(c, func() error {
		var err error
		stockResponse, err = h.inventoryClient.CheckStock(c.Request.Context(), req.ToStockCheckRequest())
		return err
	}, "check stock availability") {
		return
	}

	if !stockResponse.AllAvailable {
		http.RespondBadRequest(c, "Some items are not available in requested quantities")
		return
	}

	// Create order
	if h.HandleOrderClientOperation(c, func() error {
		_, err := h.orderClient.CreateOrder(c.Request.Context(), req.ToClientRequest())
		return err
	}, "create order") {
		http.RespondCreated(c, gin.H{"message": "Order created. Proceed to payment."}, "Order created. Proceed to payment.")
	}
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, ok := http.RequireUserID(c)
	if !ok {
		return // Error response already sent by RequireUserID
	}

	page, limit := http.GetPageLimit(c, 1, 10, 100)
	var orders []*clients.Order
	if h.HandleOrderClientOperation(c, func() error {
		var err error
		orders, err = h.orderClient.GetUserOrders(c.Request.Context(), userID, page, limit)
		return err
	}, "get user orders") {
		http.RespondSuccess(c, gin.H{"orders": orders, "page": page, "limit": limit, "total": len(orders)}, "Orders retrieved successfully")
	}
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, ok := h.RequireUserID(c)
	if !ok {
		return // Error response already sent by RequireUserID
	}

	orderID, ok := h.RequireParam(c, "id")
	if !ok {
		return // Error response already sent by RequireParam
	}

	if h.HandleOrderClientOperation(c, func() error {
		_, err := h.orderClient.GetOrder(c.Request.Context(), orderID, userID)
		return err
	}, "get order") {
		http.RespondSuccess(c, gin.H{"message": "Order retrieved successfully"}, "Order retrieved successfully")
	}
}
