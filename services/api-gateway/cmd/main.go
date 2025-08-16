package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"api-gateway/internal/clients"
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
	"api-gateway/internal/session"
	"api-gateway/internal/token"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize gRPC clients (as interfaces)
	userClient, err := clients.NewUserClient(cfg.UserServiceURL)
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer userClient.Close()

	orderClient, err := clients.NewOrderClient(cfg.OrderServiceURL)
	if err != nil {
		log.Fatalf("Failed to connect to order service: %v", err)
	}
	defer orderClient.Close()

	inventoryClient, err := clients.NewInventoryClient(cfg.InventoryServiceURL)
	if err != nil {
		log.Fatalf("Failed to connect to inventory service: %v", err)
	}
	defer inventoryClient.Close()

	paymentClient, err := clients.NewPaymentClient(cfg.PaymentServiceURL)
	if err != nil {
		log.Fatalf("Failed to connect to payment service: %v", err)
	}
	defer paymentClient.Close()

	// Initialize Redis + stores
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	sessStore := session.NewRedisSessionStore(rdb, "sess:")
	// Simplified access token minter: opaque tokens stored in Redis
	minter := token.NewSimpleMinter(sessStore, 15*60)

	// Initialize handlers
	userHandler := handlers.NewUserHandlerWithSessions(userClient, sessStore, cfg)
	authHandler := handlers.NewAuthHandlerWithDeps(sessStore, minter, cfg)
	orderHandler := handlers.NewOrderHandler(orderClient, inventoryClient, paymentClient)
	inventoryHandler := handlers.NewInventoryHandler(inventoryClient)
	paymentHandler := handlers.NewPaymentHandler(paymentClient)

	// Setup router
	router := gin.Default()

	// Trusted proxies
	trusted := os.Getenv("TRUSTED_PROXIES")
	if trusted == "" {
		trusted = "127.0.0.1,172.16.0.0/12"
	}
	var proxyList []string
	for _, p := range strings.Split(trusted, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			proxyList = append(proxyList, p)
		}
	}
	if err := router.SetTrustedProxies(proxyList); err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// JWKS env bridge for middleware
	if os.Getenv("JWKS_URL") == "" && cfg.JWKSURL != "" {
		_ = os.Setenv("JWKS_URL", cfg.JWKSURL)
	}
	if os.Getenv("JWT_ISSUER") == "" && cfg.JWTIssuer != "" {
		_ = os.Setenv("JWT_ISSUER", cfg.JWTIssuer)
	}
	if os.Getenv("JWT_AUDIENCE") == "" && cfg.JWTAudience != "" {
		_ = os.Setenv("JWT_AUDIENCE", cfg.JWTAudience)
	}

	// CORS middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// User routes
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
			users.GET("/profile", middleware.AuthRequired(), userHandler.GetProfile)
			users.PUT("/profile", middleware.AuthRequired(), userHandler.UpdateProfile)
		}

		// Auth/Session routes
		auth := api.Group("/auth")
		{
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		// Inventory routes
		inventory := api.Group("/inventory")
		{
			inventory.GET("/products", inventoryHandler.GetProducts)
			inventory.GET("/products/:id", inventoryHandler.GetProduct)
			inventory.GET("/categories", inventoryHandler.GetCategories)
		}

		// Order routes (protected)
		orders := api.Group("/orders")
		orders.Use(middleware.AuthRequired())
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.GetUserOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PUT("/:id/cancel", orderHandler.CancelOrder)
		}

		// Payment routes
		// Public informational endpoints
		api.GET("/payments/methods", paymentHandler.GetPaymentMethods)
		api.GET("/payments/test-cards", paymentHandler.GetTestCards)

		// Protected payment operations
		payments := api.Group("/payments")
		payments.Use(middleware.AuthRequired())
		{
			payments.POST("", paymentHandler.ProcessPayment)
			payments.GET("/:id", paymentHandler.GetPayment)
			payments.POST("/:id/refund", paymentHandler.RefundPayment)
		}
	}

	log.Printf("API Gateway starting on port %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}
