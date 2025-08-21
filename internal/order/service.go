package order

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type OrderService struct {
	repo   Repository
	cache  Cache
	writer Writer
	logger Logger
}

func NewOrderService(repository Repository, cache Cache, writer Writer, logger Logger) *OrderService {
	return &OrderService{repo: repository, logger: logger, cache: cache, writer: writer}
}

func (s *OrderService) SaveOrder(ctx context.Context, order Order) error {
	s.cache.Set(order)
	return s.repo.Save(ctx, order)
}

func (s *OrderService) GetOrderById(ctx context.Context, orderId string) (Order, error) {
	if order, exists := s.cache.Get(orderId); exists {
		return *order, nil
	}
	return s.repo.GetById(ctx, orderId)
}

func (s *OrderService) GetOrdersLimit(ctx context.Context, limit int) ([]Order, error) {
	if s.cache != nil {
		recent := s.cache.GetRecent(limit)
		if len(recent) > 0 {
			return recent, nil
		}
	}
	orders, err := s.repo.GetLimit(ctx, limit)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) CreateOrder(ctx context.Context) (Order, error) {
	order := s.generateRandomOrder()
	payload, err := json.Marshal(order)
	if err != nil {
		s.logger.Printf("marshal generated order: %v", err)
		return Order{}, err
	}
	msg := kafkago.Message{Value: payload}
	if err := s.writer.WriteMessages(ctx, msg); err != nil {
		return Order{}, err
	}

	return order, nil
}

func (s *OrderService) generateRandomOrder() Order {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	orderUID := fmt.Sprintf("%x", rng.Int63())
	trackNumber := fmt.Sprintf("%06x", rng.Int31())
	transaction := fmt.Sprintf("%x", rng.Int63())

	numProducts := rng.Intn(5) + 1
	products := make([]Product, 0, numProducts)
	for i := 0; i < numProducts; i++ {
		chrtID := rng.Intn(1000)
		prod := Product{
			ChrtID:      chrtID,
			TrackNumber: trackNumber,
			Price:       453,
			Rid:         fmt.Sprintf("rid-%x", rng.Int63()),
			Name:        "Mascaras",
			Sale:        30,
			Size:        "0",
			TotalPrice:  317,
			NmID:        2389212,
			Brand:       "Vivienne Sabo",
			Status:      202,
		}
		products = append(products, prod)
	}

	ord := Order{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: Payment{
			Transaction:  transaction,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Products:          products,
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now().UTC(),
		OofShard:          "1",
	}
	return ord
}
