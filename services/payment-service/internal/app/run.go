package app

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	pub "github.com/kubernetestest/ecommerce-platform/pkg/kafkaclient"
	pkglogger "github.com/kubernetestest/ecommerce-platform/pkg/logger"
	"github.com/kubernetestest/ecommerce-platform/pkg/metrics"
	"github.com/kubernetestest/ecommerce-platform/proto-go/events"
	app "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/app/services"
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/entities"
	derrors "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/errors"
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/valueobjects"
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/infra/cache"
	srv "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/infra/grpc"
	con "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/infra/kafka/consumer"
	mockproc "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/infra/processor"
	paymentmetrics "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/metrics"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/proto"
)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

	server := grpc.NewServer()
	processor := mockproc.NewMockPaymentProcessor()

	// Initialize metrics
	metricsInstance := paymentmetrics.NewPaymentMetrics()

	paymentService := app.NewPaymentService(processor, metricsInstance)
	srv.RegisterPaymentPBServer(server, paymentService, metricsInstance)

	// Kafka wiring (best-effort)
	var prod pub.Publisher
	var consSR *con.Consumer
	var consOC *con.OrderCreatedConsumer

	// Redis cache for order totals from OrderCreated
	var totalsCache cache.OrderTotalsCache
	var closers []io.Closer
	if cfg.RedisAddr != "" {
		totalsCache = cache.NewRedisOrderTotalsCache(cfg.RedisAddr, cfg.RedisDB, "order:total:")
		closers = append(closers, totalsCache)
		log.Infow("Redis cache initialized", "addr", cfg.RedisAddr, "db", cfg.RedisDB)
	}
	if cfg.KafkaBrokers != "" {
		config := pub.PublisherConfig{
			BootstrapServers: cfg.KafkaBrokers,
			ClientID:         "payment-service",
		}
		if p, err := pub.NewKafkaPublisher(config); err != nil {
			log.Warnw("kafka producer init failed", "error", err)
		} else {
			prod = p.WithLogger(pkglogger.NewZapLogger(log))
			closers = append(closers, prod)
			log.Infow("Kafka producer initialized", "brokers", cfg.KafkaBrokers)
		}
		// consume OrderCreated to cache totals
		if oc, err := con.NewOrderCreatedConsumer(cfg.KafkaBrokers, "payment-service", con.OrderCreatedHandlerFunc(func(cctx context.Context, evt *events.OrderCreated) error {
			if totalsCache != nil {
				if err := totalsCache.Set(cctx, evt.OrderId, evt.TotalAmount, evt.Currency, cfg.OrderTotalTTL); err != nil {
					log.Warnw("cache set failed", "orderID", evt.OrderId, "error", err)
				}
			}
			return nil
		})); err != nil {
			log.Warnw("kafka order-created consumer init failed", "error", err)
		} else {
			consOC = oc.WithLogger(pkglogger.NewZapLogger(log))
			closers = append(closers, consOC)
			go consOC.Run(ctx, []string{"orders.v1.order_created"})
			log.Infow("Kafka consumer started", "topic", "orders.v1.order_created")
		}

		// consume StockReserved to process payments
		if c, err := con.NewConsumer(cfg.KafkaBrokers, "payment-service", con.StockReservedHandlerFunc(func(cctx context.Context, evt *events.StockReserved) error {
			// Build amount from Redis cached order total if present
			amt := func() valueobjects.Money {
				if totalsCache != nil {
					if a, c, ok, _ := totalsCache.Get(cctx, evt.OrderId); ok {
						if m, e := valueobjects.NewMoney(a, c); e == nil {
							return m
						}
					}
				}
				m, _ := valueobjects.NewMoney(0, "USD")
				return m
			}()
			req := &app.ProcessPaymentRequest{OrderID: evt.OrderId, UserID: evt.UserId, Amount: amt, Method: entities.MethodCreditCard, CardNumber: "4111111111111111"}
			// Timeout for processing to avoid hanging
			hctx, cancel := context.WithTimeout(cctx, cfg.PaymentProcessTimeout)
			resp, err := paymentService.ProcessPayment(hctx, req)
			cancel()
			if err != nil && !errors.Is(err, derrors.ErrPaymentDeclined) {
				log.Warnw("process payment failed (technical)", "orderID", evt.OrderId, "userID", evt.UserId, "error", err)
				return nil
			}
			// publish outcome (both success and business-decline)
			if prod != nil && resp != nil && resp.Payment != nil {
				pe := &events.PaymentProcessed{OrderId: resp.Payment.OrderID, PaymentId: resp.Payment.ID, Success: resp.Success, Message: resp.Message, Amount: resp.Payment.Amount.Amount, Currency: resp.Payment.Amount.Currency, OccurredAt: time.Now().Format(time.RFC3339)}
				bytes, _ := proto.Marshal(pe)
				pctx, pcancel := context.WithTimeout(cctx, cfg.KafkaPublishTimeout)
				if perr := prod.Publish(pctx, "payments.v1.payment_processed", bytes); perr != nil {
					log.Errorw("publish PaymentProcessed failed", "orderID", pe.OrderId, "paymentID", pe.PaymentId, "userID", evt.UserId, "error", perr)
				} else {
					log.Infow("PaymentProcessed published", "orderID", pe.OrderId, "paymentID", pe.PaymentId, "userID", evt.UserId, "success", pe.Success)
				}
				pcancel()
				// best-effort cleanup of cached total
				if totalsCache != nil {
					if err := totalsCache.Del(cctx, evt.OrderId); err != nil {
						log.Warnw("cache delete failed", "orderID", evt.OrderId, "error", err)
					}
				}
			}
			return nil
		})); err != nil {
			log.Warnw("kafka consumer init failed", "error", err)
		} else {
			consSR = c.WithLogger(pkglogger.NewZapLogger(log))
			closers = append(closers, consSR)
			go consSR.Run(ctx, []string{"inventory.v1.stock_reserved"})
			log.Infow("Kafka consumer started", "topic", "inventory.v1.stock_reserved")
		}
	}

	// Start metrics server
	metricsServer := metrics.NewMetricsServer(":"+cfg.MetricsPort, logger)
	go func() {
		log.Infow("metrics server starting", "port", cfg.MetricsPort)
		if err := metricsServer.Start(); err != nil {
			log.Errorw("metrics server failed", "error", err)
		}
	}()

	// Graceful shutdown for metrics server
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Errorw("metrics server shutdown failed", "error", err)
		}
	}()

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Errorw("listen failed", "error", err)
		return err
	}
	log.Infow("payment-service starting", "port", cfg.Port)

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(lis) }()

	shutdown := func() {
		// graceful gRPC with timeout fallback
		done := make(chan struct{})
		go func() { server.GracefulStop(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			server.Stop()
		}
		for _, c := range closers {
			_ = c.Close()
		}
	}

	select {
	case <-ctx.Done():
		log.Infow("shutting down payment-service...")
		shutdown()
		return nil
	case err := <-serveErr:
		log.Errorw("serve failed", "error", err)
		shutdown()
		return err
	}
}
