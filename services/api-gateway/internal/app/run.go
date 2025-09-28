package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ecommerce-platform/pkg/jwt"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/api-gateway/internal/clients"
	"ecommerce-platform/services/api-gateway/internal/config"
	"ecommerce-platform/services/api-gateway/internal/handlers"
	"ecommerce-platform/services/api-gateway/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config, log *zap.Logger) error {
	// Logger adapter
	loggerAdapter := logger.NewZapLogger(log.Sugar())

	// Create clients
	orderClient, err := clients.NewOrderClient(cfg.Services.OrderServiceURL)
	if err != nil {
		loggerAdapter.Error("failed to create order client", "error", err)
		return fmt.Errorf("failed to create order client: %w", err)
	}
	defer orderClient.Close()

	userClient, err := clients.NewUserClient(cfg.Services.UserServiceURL)
	if err != nil {
		loggerAdapter.Error("failed to create user client", "error", err)
		return fmt.Errorf("failed to create user client: %w", err)
	}
	defer userClient.Close()

	inventoryClient, err := clients.NewInventoryClient(cfg.Services.InventoryServiceURL)
	if err != nil {
		loggerAdapter.Error("failed to create inventory client", "error", err)
		return fmt.Errorf("failed to create inventory client: %w", err)
	}
	defer inventoryClient.Close()

	// JWT manager
	jwtConfig := cfg.GetJWTConfig()
	jwtManager := jwt.NewManager(jwt.Config{
		AccessTokenSecret:  jwtConfig.AccessSecret,
		RefreshTokenSecret: jwtConfig.RefreshSecret,
		AccessTokenTTL:     jwtConfig.AccessTTL,
		RefreshTokenTTL:    jwtConfig.RefreshTTL,
		Issuer:             jwtConfig.Issuer,
		Audience:           jwtConfig.Audience,
	})

	// Handlers
	orderHandler := handlers.NewOrderHandler(orderClient)
	userHandler := handlers.NewUserHandler(userClient)
	inventoryHandler := handlers.NewInventoryHandler(inventoryClient)

	// Router setup
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// --- CORS middleware (using gin-contrib/cors) ---
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public auth
		authPublic := api.Group("/auth")
		{
			authPublic.POST("/register", userHandler.Register)
			authPublic.POST("/login", userHandler.Login)
			authPublic.POST("/refresh", userHandler.Refresh)
		}

		// Protected auth
		authProtected := api.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(jwtManager))
		{
			authProtected.POST("/logout", userHandler.Logout)
		}

		// Protected users
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			users.GET("/:id", userHandler.GetUser)
		}

		// Public inventory
		inventory := api.Group("/inventory")
		{
			inventory.GET("/products", inventoryHandler.GetProducts)
			inventory.GET("/products/:id", inventoryHandler.GetProduct)
		}

		// Protected orders
		orders := api.Group("/orders")
		orders.Use(middleware.AuthMiddleware(jwtManager))
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.GetUserOrders)
			orders.GET("/:id", orderHandler.GetOrder)
		}
	}

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		loggerAdapter.Info("starting api-gateway", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerAdapter.Error("failed to start server", "error", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	loggerAdapter.Info("shutting down api-gateway...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		loggerAdapter.Error("failed to shutdown server", "error", err)
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	loggerAdapter.Info("api-gateway stopped gracefully")
	return nil
}
