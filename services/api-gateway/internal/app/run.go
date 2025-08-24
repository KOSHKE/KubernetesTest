package app

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kubernetestest/ecommerce-platform/pkg/metrics"
	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/internal/clients"
	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/internal/config"
	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/internal/handlers"
	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/internal/middleware"

	cors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config, logger *zap.Logger) error {
	sugar := logger.Sugar()

	userClient, err := clients.NewUserClient(cfg.UserServiceURL)
	if err != nil {
		sugar.Errorw("user client connect failed", "error", err)
		return err
	}
	defer func() { _ = userClient.Close() }()
	orderClient, err := clients.NewOrderClient(cfg.OrderServiceURL)
	if err != nil {
		sugar.Errorw("order client connect failed", "error", err)
		return err
	}
	defer func() { _ = orderClient.Close() }()
	inventoryClient, err := clients.NewInventoryClient(cfg.InventoryServiceURL)
	if err != nil {
		sugar.Errorw("inventory client connect failed", "error", err)
		return err
	}
	defer func() { _ = inventoryClient.Close() }()
	paymentClient, err := clients.NewPaymentClient(cfg.PaymentServiceURL)
	if err != nil {
		sugar.Errorw("payment client connect failed", "error", err)
		return err
	}
	defer func() { _ = paymentClient.Close() }()

	userHandler := handlers.NewUserHandler(userClient)
	orderHandler := handlers.NewOrderHandler(orderClient, inventoryClient, paymentClient)
	inventoryHandler := handlers.NewInventoryHandler(inventoryClient)
	paymentHandler := handlers.NewPaymentHandler(paymentClient)

	// Initialize metrics
	promMetrics := metrics.NewPrometheusMetrics("api-gateway")
	metricsServer := metrics.NewMetricsServer(":"+cfg.MetricsPort, sugar.Desugar())

	router := gin.Default()
	// Strict CORS for prod via env list
	allowed := func() []string {
		var o []string
		for _, s := range strings.Split(cfg.FrontendOrigins, ",") {
			if s = strings.TrimSpace(s); s != "" {
				o = append(o, s)
			}
		}
		return o
	}()
	sugar.Infow("CORS configuration", "allowed_origins", allowed)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowed,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Accept", "Origin", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // Cache preflight requests for 12 hours
	}))

	// Set trusted proxies for development (can be overridden via TRUSTED_PROXIES env var)
	trusted := os.Getenv("TRUSTED_PROXIES")
	if trusted == "" {
		// In development, trust localhost and common dev IPs
		trusted = "127.0.0.1,::1,172.16.0.0/12,10.0.0.0/8"
	}
	var proxyList []string
	for _, p := range strings.Split(trusted, ",") {
		if p = strings.TrimSpace(p); p != "" {
			proxyList = append(proxyList, p)
		}
	}
	sugar.Infow("Trusted proxies configuration", "proxies", proxyList)
	if err := router.SetTrustedProxies(proxyList); err != nil {
		sugar.Errorw("set proxies failed", "error", err)
		return err
	}

	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) })

	// Add metrics endpoint to main HTTP server
	router.GET("/metrics", func(c *gin.Context) {
		// Get metrics from metrics server
		metricsMux := metricsServer.GetMux()
		metricsMux.ServeHTTP(c.Writer, c.Request)
	})

	// Add metrics middleware to track HTTP requests
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Record HTTP metrics with proper status code
		status := strconv.Itoa(c.Writer.Status())
		promMetrics.HTTPRequestsTotal(c.Request.Method, c.Request.URL.Path, status)
		promMetrics.HTTPRequestDuration(c.Request.Method, c.Request.URL.Path, duration)
	})

	api := router.Group("/api/v1")
	{
		// Auth routes (no middleware)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/logout", userHandler.Logout)
		}

		// Protected routes (with auth middleware)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
				users.PUT("/profile", userHandler.UpdateProfile)
			}

			orders := protected.Group("/orders")
			{
				orders.POST("", orderHandler.CreateOrder)
				orders.GET("", orderHandler.GetUserOrders)
				orders.GET("/:id", orderHandler.GetOrder)
			}

			payments := protected.Group("/payments")
			{
				payments.POST("", paymentHandler.ProcessPayment)
				payments.GET("/:id", paymentHandler.GetPayment)
				payments.POST("/:id/refund", paymentHandler.RefundPayment)
			}
		}

		// Public routes (no auth required)
		inventory := api.Group("/inventory")
		{
			inventory.GET("/products", inventoryHandler.GetProducts)
			inventory.GET("/products/:id", inventoryHandler.GetProduct)
			inventory.GET("/categories", inventoryHandler.GetCategories)
		}

		api.GET("/payments/methods", paymentHandler.GetPaymentMethods)
		api.GET("/payments/test-cards", paymentHandler.GetTestCards)
	}

	httpServer := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	// gracefulShutdown performs graceful shutdown for all servers
	gracefulShutdown := func(shutdownCtx context.Context) {
		sugar.Infow("shutting down api-gateway...")
		_ = httpServer.Shutdown(shutdownCtx)
		_ = metricsServer.Shutdown(shutdownCtx)
	}

	// Start metrics server
	metricsErr := make(chan error, 1)
	go func() { metricsErr <- metricsServer.Start() }()
	sugar.Infow("metrics server starting", "port", cfg.MetricsPort)

	serveErr := make(chan error, 1)
	go func() { serveErr <- httpServer.ListenAndServe() }()
	sugar.Infow("api-gateway starting", "port", cfg.Port)

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		gracefulShutdown(shutdownCtx)
		return nil
	case err := <-serveErr:
		if err != nil && err != http.ErrServerClosed {
			sugar.Errorw("http serve failed", "error", err)
			return err
		}
	case err := <-metricsErr:
		if err != nil {
			sugar.Errorw("metrics server failed", "error", err)
			return err
		}
	}

	// If we reached here, one of the servers failed with an error
	// Wait for context cancellation for graceful shutdown
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	gracefulShutdown(shutdownCtx)
	return nil
}
