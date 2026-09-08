package handler

import (
	"log/slog"
	"net/http"

	"parent-service/internal/logger"

	"github.com/gin-gonic/gin"
)

// SetupRouter настраивает роутинг сервиса-эмулятора заведения.
func SetupRouter(h *RestaurantHandler) *gin.Engine {
	r := gin.New()
	r.Use(logger.GinMiddleware(slog.Default()))
	r.Use(gin.Recovery())

	// Проверка здоровья сервиса
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Webhook для приема заказов от Core Service
	r.POST("/webhook/orders", h.HandleWebhook)

	v1 := r.Group("/api/v1")
	{
		// Просмотр полученных заказов
		v1.GET("/orders", h.GetReceivedOrders)
		// Эмуляция отправки стоп-листа в Core Service
		v1.POST("/simulate/stop-list", h.SimulateStopList)
	}

	return r
}
