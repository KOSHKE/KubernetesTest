//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	clients "ecommerce-platform/services/api-gateway/internal/clients"
	"ecommerce-platform/services/api-gateway/internal/handlers"
	testhelpers "ecommerce-platform/services/api-gateway/tests/integration/helpers"
	"ecommerce-platform/services/api-gateway/tests/mocks"
)

func TestInventoryEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockInventoryClient(ctrl)

	// subtest: GET /api/v1/inventory/products -> 200
	t.Run("GetProducts_Returns200", func(t *testing.T) {
		mockClient.
			EXPECT().
			GetProducts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&clients.ListProductsResult{
				Products: []*clients.Product{{
					ID: "p1", Name: "Product 1", PriceAmount: 1000, Currency: "USD", ImageURL: "https://example.com/p1.jpg", StockQuantity: 5,
				}},
				Total: 1,
				Page:  1,
				Limit: 10,
			}, nil)

		h := handlers.NewInventoryHandler(mockClient)
		r := testhelpers.BuildInventoryRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/products?page=1&limit=10", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})
}
