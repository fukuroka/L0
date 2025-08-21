package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"L0/internal/api"
	"L0/internal/order"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ms := &mockService{}
	h := api.NewHandler(ms)
	r := h.RegisterOrderRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestGetOrderSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ord := order.Order{OrderUID: "order-z"}
	ms := &mockService{order: ord}
	h := api.NewHandler(ms)
	r := h.RegisterOrderRouter()

	req := httptest.NewRequest(http.MethodGet, "/orders/order-z", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	var got order.Order
	if err := json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&got); err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if got.OrderUID != "order-z" {
		t.Fatalf("unexpected uid: %s", got.OrderUID)
	}
}

func TestGetOrdersWithLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orders := []order.Order{{OrderUID: "order1"}, {OrderUID: "order2"}}
	ms := &mockService{orders: orders}
	h := api.NewHandler(ms)
	r := h.RegisterOrderRouter()

	req := httptest.NewRequest(http.MethodGet, "/orders/?limit=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestCreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ord := order.Order{OrderUID: "order-new"}
	ms := &mockService{order: ord}
	h := api.NewHandler(ms)
	r := h.RegisterOrderRouter()

	req := httptest.NewRequest(http.MethodPost, "/orders/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Code)
	}
}
