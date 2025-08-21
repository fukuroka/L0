package api

import (
	"L0/internal/order"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type OrderHandler struct {
	service order.Service
}

func NewHandler(orderService order.Service) *OrderHandler {
	return &OrderHandler{service: orderService}
}

func (o *OrderHandler) RegisterOrderRouter() http.Handler {
	router := gin.Default()

	router.GET("/healthcheck", o.Health)
	router.Static("/static", "./internal/web")
	router.StaticFile("/", "./internal/web/index.html")
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	orderGroup := router.Group("/orders")
	{
		orderGroup.GET("/:id", o.GetOrder)
		orderGroup.GET("/", o.GetOrders)
		orderGroup.POST("/", o.CreateOrder)
	}

	return router
}

// Health godoc
// @Summary      Health check
// @Description  Returns service health status
// @Tags         orders
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /healthcheck [get]
func (o *OrderHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetOrder godoc
// @Summary      Get order by id
// @Description  Get order details by order UID
// @Tags         orders
// @Produce      json
// @Param        id   path      string  true  "Order UID"
// @Success      200  {object}  order.Order
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /orders/{id} [get]
func (o *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := o.service.GetOrderById(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// GetOrders godoc
// @Summary      Get list of orders
// @Description  Returns list of orders. Query param `limit` optional (default 10)
// @Tags         orders
// @Produce      json
// @Param        limit  query    int  false  "Limit of ids to return"
// @Success      200  {array}   string
// @Failure      500  {object}  map[string]string
// @Router       /orders/ [get]
func (o *OrderHandler) GetOrders(c *gin.Context) {
	q := c.Query("limit")
	limit := 10
	if q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			limit = v
		}
	}

	ids, err := o.service.GetOrdersLimit(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ids)
}

// CreateOrder godoc
// @Summary      Create random order and publish to Kafka
// @Description  Generates a random order, publishes it to Kafka and returns the created order
// @Tags         orders
// @Produce      json
// @Success      201  {object}  order.Order
// @Failure      500  {object}  map[string]string
// @Router       /orders/ [post]
func (o *OrderHandler) CreateOrder(c *gin.Context) {
	order, err := o.service.CreateOrder(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}
