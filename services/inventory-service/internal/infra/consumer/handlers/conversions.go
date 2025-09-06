package handlers

import (
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
)

// convertToStockReservationItems converts valueobjects.Item to dto.StockReservationItem
func convertToStockReservationItems(items []valueobjects.Item) []dto.StockReservationItem {
	result := make([]dto.StockReservationItem, len(items))
	for i, item := range items {
		result[i] = dto.StockReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return result
}
