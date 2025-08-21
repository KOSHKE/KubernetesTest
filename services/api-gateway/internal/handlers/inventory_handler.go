package handlers

import (
	"api-gateway/internal/clients"
	"api-gateway/pkg/http"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	http.BaseHandler
	inventoryClient clients.InventoryClient
}

func NewInventoryHandler(inventoryClient clients.InventoryClient) *InventoryHandler {
	return &InventoryHandler{inventoryClient: inventoryClient}
}

func (h *InventoryHandler) GetProducts(c *gin.Context) {
	categoryID := c.Query("category_id")
	search := c.Query("search")
	page, limit := http.GetPageLimit(c, 1, 20, 100)
	products, total, err := h.inventoryClient.GetProducts(c.Request.Context(), categoryID, page, limit, search)
	if err != nil {
		http.HandleInventoryClientError(c, err, "get products")
		return
	}
	http.RespondSuccess(c, gin.H{"products": products, "page": page, "limit": limit, "total": total}, "Products retrieved successfully")
}

func (h *InventoryHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		http.RespondBadRequest(c, "Product ID is required")
		return
	}
	product, err := h.inventoryClient.GetProduct(c.Request.Context(), productID)
	if err != nil {
		http.HandleInventoryClientError(c, err, "get product")
		return
	}
	http.RespondSuccess(c, gin.H{"product": product}, "Product retrieved successfully")
}

func (h *InventoryHandler) GetCategories(c *gin.Context) {
	activeOnly := c.Query("active_only") == "true"
	categories, err := h.inventoryClient.GetCategories(c.Request.Context(), activeOnly)
	if err != nil {
		http.HandleInventoryClientError(c, err, "get categories")
		return
	}
	http.RespondSuccess(c, gin.H{"categories": categories}, "Categories retrieved successfully")
}
