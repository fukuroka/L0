package cache

import (
	"L0/internal/order"
	"sync"
)

type CacheOrder struct {
	Mu       sync.Mutex
	Size     int
	Orders   map[string]order.Order
	OrderIds []string
}

func NewCache(size int) *CacheOrder {
	return &CacheOrder{Size: size, Orders: make(map[string]order.Order), OrderIds: make([]string, 0, size)}
}

func (cache *CacheOrder) Set(instance order.Order) {
	cache.Mu.Lock()

	defer cache.Mu.Unlock()

	cache.Orders[instance.OrderUID] = instance

	if len(cache.OrderIds) < cache.Size {
		cache.OrderIds = append([]string{instance.OrderUID}, cache.OrderIds...)
	} else {
		if cache.Size > 0 {
			delete(cache.Orders, cache.OrderIds[len(cache.OrderIds)-1])
			cache.OrderIds = append([]string{instance.OrderUID}, cache.OrderIds[:cache.Size-1]...)
		}
	}
}

func (cache *CacheOrder) Get(key string) (*order.Order, bool) {
	cache.Mu.Lock()
	defer cache.Mu.Unlock()
	v, ok := cache.Orders[key]
	if !ok {
		return nil, false
	}
	return &v, true
}

func (cache *CacheOrder) Load(orders []order.Order) {
	for _, order := range orders {
		cache.Set(order)
	}
}

func (cache *CacheOrder) GetRecent(limit int) []order.Order {
	cache.Mu.Lock()
	defer cache.Mu.Unlock()
	n := limit
	if n > len(cache.OrderIds) {
		n = len(cache.OrderIds)
	}
	res := make([]order.Order, 0, n)
	for i := 0; i < n; i++ {
		id := cache.OrderIds[i]
		if o, ok := cache.Orders[id]; ok {
			res = append(res, o)
		}
	}
	return res
}
