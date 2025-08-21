package order

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

type Cache interface {
	Set(instance Order)
	Get(key string) (*Order, bool)
	GetRecent(limit int) []Order
}

type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
}

type Repository interface {
	Save(ctx context.Context, order Order) error
	GetById(ctx context.Context, orderId string) (Order, error)
	GetLimit(ctx context.Context, limit int) ([]Order, error)
}

type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafkago.Message) error
}

type Service interface {
	SaveOrder(ctx context.Context, order Order) error
	GetOrderById(ctx context.Context, orderId string) (Order, error)
	GetOrdersLimit(ctx context.Context, limit int) ([]Order, error)
	CreateOrder(ctx context.Context) (Order, error)
}
