package handlers

import (
	"context"
	"strconv"
	"time"

	"ecommerce-platform/services/api-gateway/internal/clients"

	"github.com/gin-gonic/gin"
)

const inventoryDefaultRPCTimeout = 3 * time.Second

type InventoryHandler struct {
	inventoryClient clients.InventoryClient
}

func NewInventoryHandler(inventoryClient clients.InventoryClient) *InventoryHandler {
	return &InventoryHandler{inventoryClient: inventoryClient}
}

// GET /api/v1/inventory/products
func (h *InventoryHandler) GetProducts(c *gin.Context) {
	categoryID := c.Query("category_id")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), inventoryDefaultRPCTimeout)
	defer cancel()
	res, err := h.inventoryClient.GetProducts(ctx, categoryID, page, limit, search)
	if !handleGRPCError(c, err) {
		return
	}
	// Shape to frontend contract
	products := make([]gin.H, 0, len(res.Products))
	for _, p := range res.Products {
		products = append(products, gin.H{
			"id":             p.ID,
			"name":           p.Name,
			"price":          gin.H{"amount": p.PriceAmount, "currency": p.Currency},
			"stock_quantity": p.StockQuantity,
			"image_url":      p.ImageURL,
		})
	}
	respondSuccess(c, gin.H{"products": products, "total": res.Total, "page": res.Page, "limit": res.Limit}, "Products retrieved successfully")
}

// GET /api/v1/inventory/products/:id
func (h *InventoryHandler) GetProduct(c *gin.Context) {
	id, ok := requireParam(c, "id")
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), inventoryDefaultRPCTimeout)
	defer cancel()
	p, err := h.inventoryClient.GetProduct(ctx, id)
	if !handleGRPCError(c, err) {
		return
	}
	respondSuccess(c, gin.H{
		"id":             p.ID,
		"name":           p.Name,
		"price":          gin.H{"amount": p.PriceAmount, "currency": p.Currency},
		"stock_quantity": p.StockQuantity,
		"image_url":      p.ImageURL,
	}, "Product retrieved successfully")
}

// no categories endpoint
