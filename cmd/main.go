package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "L0/docs"
	"L0/internal/api"
	"L0/internal/cache"
	"L0/internal/config"
	"L0/internal/db"
	"L0/internal/kafka"
	"L0/internal/order"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	log.Printf("config loaded: http_port=%s db=%s:%s/%s", cfg.HTTPPort, cfg.DB.Host, cfg.DB.Port, cfg.DB.DbName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	pool, orderServ, kafkaConsumer, router, err := setup(ctx, cfg)
	if err != nil {
		log.Fatalf("setup failed: %v", err)
	}

	server := http.Server{Addr: ":" + cfg.HTTPPort, Handler: router}

	defer func() {
		pool.Close()
		log.Println("db pool closed")
	}()

	startHTTPServer(&server)
	startKafkaConsumer(ctx, kafkaConsumer, orderServ)

	log.Println("service started")

	<-sigCh
	log.Println("shutdown signal received, shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	cancel()
	log.Println("service stopped")
}

func setup(ctx context.Context, cfg config.Config) (pool *pgxpool.Pool, orderServ *order.OrderService, kafkaConsumer *kafkago.Reader, router http.Handler, err error) {
	p, err := db.NewClient(ctx, cfg.DB)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	cons := kafka.NewConsumer(cfg.Kafka)
	wr := kafka.NewWriter(cfg.Kafka)

	c := cache.NewCache(cfg.CacheSize)
	logger := log.Default()
	orderRepo := order.NewOrderRepository(p, logger)
	s := order.NewOrderService(orderRepo, c, wr, logger)

	handler := api.NewHandler(s)
	r := handler.RegisterOrderRouter()

	lastOrders, err := s.GetOrdersLimit(ctx, cfg.CacheSize)
	if err != nil {
		p.Close()
		return nil, nil, nil, nil, err
	}
	c.Load(lastOrders)

	return p, s, cons, r, nil
}

func startHTTPServer(server *http.Server) {
	go func() {
		log.Printf("http server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()
}

func startKafkaConsumer(ctx context.Context, reader *kafkago.Reader, svc *order.OrderService) {
	go kafka.RunConsumer(ctx, reader, svc)
}
