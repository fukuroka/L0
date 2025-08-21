package test

import (
	"context"
	"errors"
	"testing"

	"L0/internal/order"
)

func TestSaveOrderCallsRepoAndCache(t *testing.T) {
	repo := &mockRepo{}
	cache := &mockCache{}
	writer := &writerRec{}
	svc := order.NewOrderService(repo, cache, writer, nil)

	o := order.Order{OrderUID: "some-order"}
	if err := svc.SaveOrder(context.Background(), o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cache.Get("some-order"); !ok {
		t.Fatalf("expected order in cache")
	}
}

func TestGetOrderByIdFromCache(t *testing.T) {
	repo := &mockRepo{getErr: errors.New("should not be called")}
	cache := &mockCache{store: map[string]order.Order{"order-cache": {OrderUID: "order-cache"}}}
	writer := &writerRec{}
	svc := order.NewOrderService(repo, cache, writer, nil)

	got, err := svc.GetOrderById(context.Background(), "order-cache")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.OrderUID != "order-cache" {
		t.Fatalf("unexpected order uid: %s", got.OrderUID)
	}
}

func TestGetOrdersLimitCallsRepo(t *testing.T) {
	orders := []order.Order{{OrderUID: "order1"}, {OrderUID: "order2"}}
	repo := &mockRepo{orders: orders}
	cache := &mockCache{}
	writer := &writerRec{}
	svc := order.NewOrderService(repo, cache, writer, nil)

	res, err := svc.GetOrdersLimit(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(res))
	}
}

func TestCreateOrderWritesToWriter(t *testing.T) {
	repo := &mockRepo{}
	cache := &mockCache{}
	w := &writerRec{}
	svc := order.NewOrderService(repo, cache, w, nil)

	ord, err := svc.CreateOrder(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ord.OrderUID == "" {
		t.Fatalf("expected generated order uid")
	}
	if len(w.last) == 0 {
		t.Fatalf("expected writer to receive payload")
	}
}
