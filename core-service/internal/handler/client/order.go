package client

import (
	"net/http"

	"core-service/internal/handler"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrderHandler обрабатывает запросы клиентов по заказам.
type OrderHandler struct {
	svc service.OrderService
}

// NewOrderHandler создает экземпляр OrderHandler.
func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// CreateOrderItemRequest запрос на позицию в заказе.
type CreateOrderItemRequest struct {
	MenuItemID uuid.UUID `json:"menu_item_id" binding:"required"`
	Quantity   int       `json:"quantity" binding:"required,gt=0"`
}

// CreateOrderRequest тело запроса создания заказа.
type CreateOrderRequest struct {
	RestaurantID    uuid.UUID                `json:"restaurant_id" binding:"required"`
	CustomerName    string                   `json:"customer_name" binding:"required"`
	CustomerPhone   string                   `json:"customer_phone" binding:"required"`
	DeliveryAddress string                   `json:"delivery_address" binding:"required"`
	Items           []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CancelOrderRequest запрос на отмену заказа клиентом.
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// CreateOrder создает заказ с валидацией стоп-листов и расчетом стоимости.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Error:   "invalid request body",
			Details: []string{err.Error()},
		})
		return
	}

	var userID *uuid.UUID
	if userIDStr := c.GetHeader("X-User-ID"); userIDStr != "" {
		parsedUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, handler.ErrorResponse{
				Error:   "invalid X-User-ID header uuid",
				Details: []string{err.Error()},
			})
			return
		}
		userID = &parsedUID
	}

	items := make([]service.OrderItemInput, len(req.Items))
	for i, it := range req.Items {
		items[i] = service.OrderItemInput{
			MenuItemID: it.MenuItemID,
			Quantity:   it.Quantity,
		}
	}

	order, err := h.svc.CreateOrder(c.Request.Context(), service.CreateOrderDTO{
		UserID:          userID,
		RestaurantID:    req.RestaurantID,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		DeliveryAddress: req.DeliveryAddress,
		Items:           items,
	})
	if err != nil {
		handler.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrder возвращает данные заказа, состав блюд и полную историю смены статусов.
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid order uuid"})
		return
	}

	order, err := h.svc.GetOrder(c.Request.Context(), id)
	if err != nil {
		handler.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, order)
}

// CancelOrder отменяет заказ, если он еще не перешел в статус приготовления (COOKING).
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid order uuid"})
		return
	}

	var req CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Error:   "invalid request body",
			Details: []string{err.Error()},
		})
		return
	}

	if err := h.svc.CancelOrder(c.Request.Context(), id, req.Reason); err != nil {
		handler.RespondError(c, err)
		return
	}

	handler.RespondSuccess(c, "order successfully cancelled")
}
