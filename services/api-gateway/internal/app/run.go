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

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config, log *zap.Logger) error {
	// Create logger adapter
	loggerAdapter := logger.NewZapLogger(log.Sugar())

	// Create order client
	orderClient, err := clients.NewOrderClient(cfg.Services.OrderServiceURL)
	if err != nil {
		loggerAdapter.Error("failed to create order client", "error", err)
		return fmt.Errorf("failed to create order client: %w", err)
	}
	defer func() { _ = orderClient.Close() }()

	// Create JWT manager
	jwtConfig := cfg.GetJWTConfig()
	jwtManager := jwt.NewManager(jwt.Config{
		AccessTokenSecret:  jwtConfig.AccessSecret,
		RefreshTokenSecret: jwtConfig.RefreshSecret,
		AccessTokenTTL:     jwtConfig.AccessTTL,
		RefreshTokenTTL:    jwtConfig.RefreshTTL,
		Issuer:             jwtConfig.Issuer,
		Audience:           jwtConfig.Audience,
	})

	// Create handlers
	orderHandler := handlers.NewOrderHandler(orderClient)

	// Setup router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes - WITH AUTHENTICATION!
	api := router.Group("/api/v1")
	{
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

	// Start server in goroutine
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
