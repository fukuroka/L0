package test

import (
	"testing"

	"L0/internal/cache"
	"L0/internal/order"
)

func makeSampleOrder(id string) order.Order {
	return order.Order{OrderUID: id}
}

func TestCacheSetGetLoad(t *testing.T) {
	c := cache.NewCache(2)
	ord1 := makeSampleOrder("order1")
	ord2 := makeSampleOrder("order2")
	ord3 := makeSampleOrder("order3")

	c.Set(ord1)
	if got, ok := c.Get("order1"); !ok || got.OrderUID != "order1" {
		t.Fatalf("expected to find order order1 in cache")
	}

	c.Set(ord2)
	if got, ok := c.Get("order2"); !ok || got.OrderUID != "order2" {
		t.Fatalf("expected to find order order2 in cache")
	}

	c.Set(ord3)
	if _, ok := c.Get("order1"); ok {
		t.Fatalf("expected order order1 to be evicted")
	}
	if got, ok := c.Get("order3"); !ok || got.OrderUID != "order3" {
		t.Fatalf("expected to find order order3 in cache")
	}
}

func TestCacheLoad(t *testing.T) {
	c := cache.NewCache(5)
	orders := []order.Order{makeSampleOrder("order1"), makeSampleOrder("order2"), makeSampleOrder("order3")}
	c.Load(orders)
	for _, o := range orders {
		if got, ok := c.Get(o.OrderUID); !ok || got.OrderUID != o.OrderUID {
			t.Fatalf("expected order %s in cache", o.OrderUID)
		}
	}
}
