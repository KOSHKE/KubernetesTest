package app

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"api-gateway/internal/clients"
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
	"api-gateway/internal/session"
	"api-gateway/internal/token"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer rdb.Close()

	sessStore := session.NewRedisSessionStore(rdb, "sess:")
	minter := token.NewSimpleMinter(sessStore, 15*60)

	userHandler := handlers.NewUserHandlerWithSessions(userClient, sessStore, (*config.Config)(cfg))
	authHandler := handlers.NewAuthHandlerWithDeps(sessStore, minter, (*config.Config)(cfg))
	orderHandler := handlers.NewOrderHandler(orderClient, inventoryClient, paymentClient)
	inventoryHandler := handlers.NewInventoryHandler(inventoryClient)
	paymentHandler := handlers.NewPaymentHandler(paymentClient)

	router := gin.Default()

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

	if os.Getenv("JWKS_URL") == "" && cfg.JWKSURL != "" {
		_ = os.Setenv("JWKS_URL", cfg.JWKSURL)
	}
	if os.Getenv("JWT_ISSUER") == "" && cfg.JWTIssuer != "" {
		_ = os.Setenv("JWT_ISSUER", cfg.JWTIssuer)
	}
	if os.Getenv("JWT_AUDIENCE") == "" && cfg.JWTAudience != "" {
		_ = os.Setenv("JWT_AUDIENCE", cfg.JWTAudience)
	}

	router.Use(middleware.CORS())
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) })

	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
			users.GET("/profile", middleware.AuthRequired(), userHandler.GetProfile)
			users.PUT("/profile", middleware.AuthRequired(), userHandler.UpdateProfile)
		}
		auth := api.Group("/auth")
		{
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}
		inventory := api.Group("/inventory")
		{
			inventory.GET("/products", inventoryHandler.GetProducts)
			inventory.GET("/products/:id", inventoryHandler.GetProduct)
			inventory.GET("/categories", inventoryHandler.GetCategories)
		}
		orders := api.Group("/orders")
		orders.Use(middleware.AuthRequired())
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.GetUserOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PUT("/:id/cancel", orderHandler.CancelOrder)
		}
		api.GET("/payments/methods", paymentHandler.GetPaymentMethods)
		api.GET("/payments/test-cards", paymentHandler.GetTestCards)
		payments := api.Group("/payments")
		payments.Use(middleware.AuthRequired())
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
