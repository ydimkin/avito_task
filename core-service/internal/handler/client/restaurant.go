package client

import (
	"net/http"
	"strconv"

	"core-service/internal/handler"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RestaurantHandler обработчик запросов витрины заведений.
type RestaurantHandler struct {
	svc service.RestaurantService
}

// NewRestaurantHandler создает экземпляр RestaurantHandler.
func NewRestaurantHandler(svc service.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{svc: svc}
}

// ListRestaurants возвращает список активных ресторанов с поддержкой поиска и пагинации.
func (h *RestaurantHandler) ListRestaurants(c *gin.Context) {
	search := c.Query("search")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	restaurants, err := h.svc.List(c.Request.Context(), search, limit, offset)
	if err != nil {
		handler.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, restaurants)
}

// GetRestaurant возвращает детальную информацию о конкретном заведении.
func (h *RestaurantHandler) GetRestaurant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid restaurant uuid"})
		return
	}

	restaurant, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handler.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, restaurant)
}
