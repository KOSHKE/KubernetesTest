package handlers

import (
	"net/http"
	"strconv"

	"api-gateway/internal/clients"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	inventoryClient clients.InventoryClient
}

func NewInventoryHandler(inventoryClient clients.InventoryClient) *InventoryHandler {
	return &InventoryHandler{inventoryClient: inventoryClient}
}

func (h *InventoryHandler) GetProducts(c *gin.Context) {
	categoryID := c.Query("category_id")
	search := c.Query("search")
	page := int32(1)
	limit := int32(20)
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
	products, total, err := h.inventoryClient.GetProducts(c.Request.Context(), categoryID, page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products, "page": page, "limit": limit, "total": total})
}

func (h *InventoryHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}
	product, err := h.inventoryClient.GetProduct(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": product})
}

func (h *InventoryHandler) GetCategories(c *gin.Context) {
	activeOnly := c.Query("active_only") == "true"
	categories, err := h.inventoryClient.GetCategories(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
