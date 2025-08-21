package test

import (
	"L0/internal/order"
	"context"
	"errors"

	kafkago "github.com/segmentio/kafka-go"
)

type mockService struct {
	getErr error
	order  order.Order
	orders []order.Order
}

func (m *mockService) SaveOrder(ctx context.Context, o order.Order) error { return nil }
func (m *mockService) GetOrderById(ctx context.Context, id string) (order.Order, error) {
	if m.getErr != nil {
		return order.Order{}, m.getErr
	}
	return m.order, nil
}
func (m *mockService) GetOrdersLimit(ctx context.Context, limit int) ([]order.Order, error) {
	return m.orders, nil
}
func (m *mockService) CreateOrder(ctx context.Context) (order.Order, error) { return m.order, nil }

type mockRepo struct {
	saveErr error
	getErr  error
	orders  []order.Order
}

func (m *mockRepo) Save(ctx context.Context, o order.Order) error { return m.saveErr }
func (m *mockRepo) GetById(ctx context.Context, id string) (order.Order, error) {
	if m.getErr != nil {
		return order.Order{}, m.getErr
	}
	for _, o := range m.orders {
		if o.OrderUID == id {
			return o, nil
		}
	}
	return order.Order{}, errors.New("order not found")
}
func (m *mockRepo) GetLimit(ctx context.Context, limit int) ([]order.Order, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if limit > len(m.orders) {
		limit = len(m.orders)
	}
	return m.orders[:limit], nil
}

type mockCache struct {
	store map[string]order.Order
}

func (m *mockCache) Set(o order.Order) {
	if m.store == nil {
		m.store = map[string]order.Order{}
	}
	m.store[o.OrderUID] = o
}
func (m *mockCache) Get(key string) (*order.Order, bool) {
	v, ok := m.store[key]
	if !ok {
		return nil, false
	}
	return &v, true
}

func (m *mockCache) GetRecent(limit int) []order.Order {
	if m.store == nil || limit <= 0 {
		return []order.Order{}
	}
	res := make([]order.Order, 0, limit)
	for _, v := range m.store {
		res = append(res, v)
		if len(res) >= limit {
			break
		}
	}
	return res
}

type writerRec struct {
	last []byte
	err  error
}

func (w *writerRec) WriteMessages(ctx context.Context, msgs ...kafkago.Message) error {
	if len(msgs) > 0 {
		w.last = msgs[0].Value
	}
	return w.err
}
