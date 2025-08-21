package kafka

import (
	"context"
	"encoding/json"
	"time"

	"L0/internal/config"
	"L0/internal/order"
	"log"

	"github.com/segmentio/kafka-go"
)

func NewConsumer(cfg config.KafkaConf) *kafka.Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		Topic:       cfg.Topic,
		GroupID:     cfg.GroupID,
		StartOffset: cfg.Offset,
		MaxWait:     100 * time.Millisecond,
	})
	return reader
}

func RunConsumer(ctx context.Context, reader *kafka.Reader, svc *order.OrderService) {
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("kafka fetch error: %v", err)
			break
		}
		var ord order.Order
		if err := json.Unmarshal(m.Value, &ord); err != nil {
			log.Printf("unmarshal error: %v; payload=%s", err, string(m.Value))
			continue
		}
		if err := svc.SaveOrder(ctx, ord); err != nil {
			log.Printf("save order error (order_uid=%s): %v", ord.OrderUID, err)
			continue
		}
		if err := reader.CommitMessages(ctx, m); err != nil {
			log.Printf("commit offset error: %v", err)
		}
	}
}

func NewWriter(cfg config.KafkaConf) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
	})
}
