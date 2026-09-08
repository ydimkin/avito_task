package client

import (
	"net/http"

	"core-service/internal/handler"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MenuHandler обрабатывает запросы просмотра меню.
type MenuHandler struct {
	svc service.MenuService
}

// NewMenuHandler создает экземпляр MenuHandler.
func NewMenuHandler(svc service.MenuService) *MenuHandler {
	return &MenuHandler{svc: svc}
}

// GetMenu возвращает меню ресторана, сгруппированное по категориям блюд.
func (h *MenuHandler) GetMenu(c *gin.Context) {
	idStr := c.Param("id")
	restaurantID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid restaurant uuid"})
		return
	}

	menu, err := h.svc.GetMenu(c.Request.Context(), restaurantID)
	if err != nil {
		handler.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, menu)
}
