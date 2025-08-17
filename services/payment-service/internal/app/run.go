package app

import (
	"context"
	"net"
	"time"

	app "payment-service/internal/app/services"
	"payment-service/internal/domain/entities"
	"payment-service/internal/domain/valueobjects"
	srv "payment-service/internal/grpc"
	"payment-service/internal/infra/kafka"
	mockproc "payment-service/internal/infra/processor"
	events "proto-go/events"
	pb "proto-go/payment"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

	server := grpc.NewServer()
	processor := mockproc.NewMockPaymentProcessor()
	paymentService := app.NewPaymentService(processor)
	pb.RegisterPaymentServiceServer(server, srv.NewPaymentServer(paymentService))

	// Kafka wiring (best-effort)
	var prod *kafka.Producer
	if brokers := getEnv("KAFKA_BROKERS", "kafka:9092"); brokers != "" {
		if p, err := kafka.NewProducer(brokers, "payments.v1.payment_processed"); err == nil {
			prod = p
			defer p.Close()
		}
		if cons, err := kafka.NewConsumer(brokers, "payment-service", kafka.StockReservedHandlerFunc(func(cctx context.Context, evt *events.StockReserved) error {
			// Map to service request (mock minimal values)
			amt, _ := valueobjects.NewMoney(0, "USD")
			req := &app.ProcessPaymentRequest{OrderID: evt.OrderId, UserID: "", Amount: amt, Method: entities.MethodCreditCard, CardNumber: "4111111111111111"}
			resp, _ := paymentService.ProcessPayment(cctx, req)
			if prod != nil && resp != nil && resp.Payment != nil {
				pe := &events.PaymentProcessed{OrderId: resp.Payment.OrderID, PaymentId: resp.Payment.ID, Success: resp.Success, Message: resp.Message, Amount: resp.Payment.Amount.Amount, Currency: resp.Payment.Amount.Currency, OccurredAt: time.Now().Format(time.RFC3339)}
				_ = prod.PublishPaymentProcessed(cctx, pe)
			}
			return nil
		})); err == nil {
			defer cons.Close()
			go cons.Run(ctx, "inventory.v1.stock_reserved")
		}
	}

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

	select {
	case <-ctx.Done():
		log.Infow("shutting down payment-service...")
		server.GracefulStop()
		time.Sleep(100 * time.Millisecond)
		return nil
	case err := <-serveErr:
		log.Errorw("serve failed", "error", err)
		return err
	}
}
