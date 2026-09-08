package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"parent-service/internal/service"
)

// RestaurantHandler хендлер эмулятора заведения.
type RestaurantHandler struct {
	simulator service.KitchenSimulator
}

// NewRestaurantHandler создает экземпляр RestaurantHandler.
func NewRestaurantHandler(simulator service.KitchenSimulator) *RestaurantHandler {
	return &RestaurantHandler{simulator: simulator}
}

// HandleWebhook принимает входящие заказы от Core Service.
func (h *RestaurantHandler) HandleWebhook(c *gin.Context) {
	var order service.ReceivedOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload", "details": err.Error()})
		return
	}

	h.simulator.HandleIncomingOrder(c.Request.Context(), order)
	c.JSON(http.StatusOK, gin.H{"status": "accepted"})
}

// GetReceivedOrders возвращает список заказов, полученных эмулятором.
func (h *RestaurantHandler) GetReceivedOrders(c *gin.Context) {
	c.JSON(http.StatusOK, h.simulator.GetOrders())
}

// SimulateStopListRequest тело запроса для симуляции стоп-листа.
type SimulateStopListRequest struct {
	RestaurantID string `json:"restaurant_id" binding:"required"`
	ItemID       string `json:"item_id" binding:"required"`
	IsAvailable  *bool  `json:"is_available" binding:"required"`
}

// SimulateStopList эмулирует отправку стоп-листа в Core Service.
func (h *RestaurantHandler) SimulateStopList(c *gin.Context) {
	var req SimulateStopListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	if err := h.simulator.TriggerStopList(c.Request.Context(), req.RestaurantID, req.ItemID, *req.IsAvailable); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stop-list update dispatched"})
}
