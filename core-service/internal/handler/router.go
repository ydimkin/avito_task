package handler

import (
	"log/slog"
	"net/http"

	"core-service/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// RestaurantHandler описывает интерфейс обработчика витрины заведений.
type RestaurantHandler interface {
	ListRestaurants(c *gin.Context)
	GetRestaurant(c *gin.Context)
}

// MenuHandler описывает интерфейс обработчика меню.
type MenuHandler interface {
	GetMenu(c *gin.Context)
}

// OrderHandler описывает интерфейс обработчика заказов клиента.
type OrderHandler interface {
	CreateOrder(c *gin.Context)
	GetOrder(c *gin.Context)
	CancelOrder(c *gin.Context)
}

// PartnerHandler описывает интерфейс обработчика партнерского контура.
type PartnerHandler interface {
	SyncMenu(c *gin.Context)
	SetItemAvailability(c *gin.Context)
	SetItemStock(c *gin.Context)
	ListOrders(c *gin.Context)
	UpdateOrderStatus(c *gin.Context)
}

// RouterConfig содержит хендлеры, необходимые для построения маршрутизации приложения.
type RouterConfig struct {
	RestaurantHandler RestaurantHandler
	MenuHandler       MenuHandler
	OrderHandler      OrderHandler
	PartnerHandler    PartnerHandler
}

// SetupRouter настраивает и возвращает HTTP-движок Gin с зарегистрированными эндпоинтами.
func SetupRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(logger.GinMiddleware(slog.Default()))
	r.Use(gin.Recovery())

	// Проверка работоспособности сервиса
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger UI и OpenAPI спецификация
	RegisterSwagger(r)

	v1 := r.Group("/api/v1")
	{
		// Клиентские эндпоинты
		restaurants := v1.Group("/restaurants")
		{
			restaurants.GET("", cfg.RestaurantHandler.ListRestaurants)
			restaurants.GET("/:id", cfg.RestaurantHandler.GetRestaurant)
			restaurants.GET("/:id/menu", cfg.MenuHandler.GetMenu)
		}

		orders := v1.Group("/orders")
		{
			orders.POST("", cfg.OrderHandler.CreateOrder)
			orders.GET("/:id", cfg.OrderHandler.GetOrder)
			orders.POST("/:id/cancel", cfg.OrderHandler.CancelOrder)
		}

		// Партнерские эндпоинты
		partner := v1.Group("/partner")
		{
			partner.POST("/restaurants/:id/menu", cfg.PartnerHandler.SyncMenu)
			partner.PATCH("/restaurants/:id/items/:itemId/availability", cfg.PartnerHandler.SetItemAvailability)
			partner.PATCH("/restaurants/:id/items/:itemId/stock", cfg.PartnerHandler.SetItemStock)
			partner.GET("/restaurants/:id/orders", cfg.PartnerHandler.ListOrders)
			partner.PATCH("/orders/:id/status", cfg.PartnerHandler.UpdateOrderStatus)
		}
	}

	return r
}
