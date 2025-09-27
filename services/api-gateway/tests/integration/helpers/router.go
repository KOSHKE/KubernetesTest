//go:build integration

package helpers

import (
	"github.com/gin-gonic/gin"

	"ecommerce-platform/services/api-gateway/internal/handlers"
)

// BuildInventoryRouter builds minimal router for inventory endpoints
func BuildInventoryRouter(h *handlers.InventoryHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	inventory := api.Group("/inventory")
	inventory.GET("/products", h.GetProducts)
	inventory.GET("/products/:id", h.GetProduct)
	return r
}
