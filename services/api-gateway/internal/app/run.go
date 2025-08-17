package app

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"api-gateway/internal/clients"
	"api-gateway/internal/handlers"

	cors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	sugar := logger.Sugar()

	userClient, err := clients.NewUserClient(cfg.UserServiceURL)
	if err != nil {
		sugar.Errorw("user client connect failed", "error", err)
		return err
	}
	defer userClient.Close()
	orderClient, err := clients.NewOrderClient(cfg.OrderServiceURL)
	if err != nil {
		sugar.Errorw("order client connect failed", "error", err)
		return err
	}
	defer orderClient.Close()
	inventoryClient, err := clients.NewInventoryClient(cfg.InventoryServiceURL)
	if err != nil {
		sugar.Errorw("inventory client connect failed", "error", err)
		return err
	}
	defer inventoryClient.Close()
	paymentClient, err := clients.NewPaymentClient(cfg.PaymentServiceURL)
	if err != nil {
		sugar.Errorw("payment client connect failed", "error", err)
		return err
	}
	defer paymentClient.Close()

	userHandler := handlers.NewUserHandler(userClient)
	orderHandler := handlers.NewOrderHandler(orderClient, inventoryClient, paymentClient)
	inventoryHandler := handlers.NewInventoryHandler(inventoryClient)
	paymentHandler := handlers.NewPaymentHandler(paymentClient)

	router := gin.Default()
	router.Use(cors.Default())

	trusted := os.Getenv("TRUSTED_PROXIES")
	if trusted == "" {
		trusted = "127.0.0.1,172.16.0.0/12"
	}
	var proxyList []string
	for _, p := range strings.Split(trusted, ",") {
		if p = strings.TrimSpace(p); p != "" {
			proxyList = append(proxyList, p)
		}
	}
	if err := router.SetTrustedProxies(proxyList); err != nil {
		sugar.Errorw("set proxies failed", "error", err)
		return err
	}

	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) })

	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
		}
		inventory := api.Group("/inventory")
		{
			inventory.GET("/products", inventoryHandler.GetProducts)
			inventory.GET("/products/:id", inventoryHandler.GetProduct)
			inventory.GET("/categories", inventoryHandler.GetCategories)
		}
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.GetUserOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PUT("/:id/cancel", orderHandler.CancelOrder)
		}
		api.GET("/payments/methods", paymentHandler.GetPaymentMethods)
		api.GET("/payments/test-cards", paymentHandler.GetTestCards)
		payments := api.Group("/payments")
		{
			payments.POST("", paymentHandler.ProcessPayment)
			payments.GET("/:id", paymentHandler.GetPayment)
			payments.POST("/:id/refund", paymentHandler.RefundPayment)
		}
	}

	httpServer := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	serveErr := make(chan error, 1)
	go func() { serveErr <- httpServer.ListenAndServe() }()
	sugar.Infow("api-gateway starting", "port", cfg.Port)

	select {
	case <-ctx.Done():
		sugar.Infow("shutting down api-gateway...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		return nil
	case err := <-serveErr:
		if err != nil && err != http.ErrServerClosed {
			sugar.Errorw("http serve failed", "error", err)
			return err
		}
		return nil
	}
}
