package partner

import (
	"net/http"

	"core-service/internal/domain"
	"core-service/internal/handler"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler обрабатывает запросы от заведений-партнеров.
type Handler struct {
	menuService  service.MenuService
	orderService service.OrderService
}

// NewHandler создает экземпляр Handler.
func NewHandler(menuService service.MenuService, orderService service.OrderService) *Handler {
	return &Handler{
		menuService:  menuService,
		orderService: orderService,
	}
}

// getPartnerRestaurantID извлекает ID ресторана из X-Partner-ID / X-Restaurant-ID header или URL параметра.
func getPartnerRestaurantID(c *gin.Context) (uuid.UUID, error) {
	pathIDStr := c.Param("id")
	headerIDStr := c.GetHeader("X-Partner-ID")
	if headerIDStr == "" {
		headerIDStr = c.GetHeader("X-Restaurant-ID")
	}

	// Если переданы оба, проверяем их соответствие во избежание подмены
	if headerIDStr != "" && pathIDStr != "" && headerIDStr != pathIDStr {
		return uuid.Nil, domain.ErrPartnerIDMismatch
	}

	target := pathIDStr
	if target == "" {
		target = headerIDStr
	}
	if target == "" {
		return uuid.Nil, domain.ErrRestaurantIDRequired
	}

	return uuid.Parse(target)
}

// UpdateAvailabilityRequest запрос изменения доступности блюда в меню.
type UpdateAvailabilityRequest struct {
	IsAvailable *bool `json:"is_available" binding:"required"`
}

// UpdateStockRequest запрос изменения количественного остатка блюда.
type UpdateStockRequest struct {
	StockQuantity *int `json:"stock_quantity" binding:"required,gte=0"`
}

// UpdateOrderStatusRequest запрос изменения статуса заказа.
type UpdateOrderStatusRequest struct {
	Status  domain.OrderStatus `json:"status" binding:"required"`
	Comment *string            `json:"comment"`
}

// SyncMenuItemDTO элемент меню при синхронизации.
type SyncMenuItemDTO struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"required,gte=0"`
	StockQuantity int     `json:"stock_quantity"`
	IsAvailable   bool    `json:"is_available"`
}

// SyncMenuCategoryDTO категория меню при синхронизации.
type SyncMenuCategoryDTO struct {
	Name      string            `json:"name" binding:"required"`
	SortOrder int               `json:"sort_order"`
	Items     []SyncMenuItemDTO `json:"items" binding:"required,dive"`
}

// SyncMenuRequest пакетный запрос выгрузки нового меню.
type SyncMenuRequest struct {
	Categories []SyncMenuCategoryDTO `json:"categories" binding:"required,dive"`
}

// SetItemAvailability включает или выключает блюдо в меню заведения.
func (h *Handler) SetItemAvailability(c *gin.Context) {
	restaurantID, err := getPartnerRestaurantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid restaurant identification", Details: []string{err.Error()}})
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid item uuid"})
		return
	}

	var req UpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid request body", Details: []string{err.Error()}})
		return
	}

	if err := h.menuService.SetAvailability(c.Request.Context(), restaurantID, itemID, *req.IsAvailable); err != nil {
		handler.RespondError(c, err)
		return
	}

	handler.RespondSuccess(c, "item availability successfully updated")
}

// SetItemStock устанавливает точное количество порций/товара на складе.
func (h *Handler) SetItemStock(c *gin.Context) {
	restaurantID, err := getPartnerRestaurantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid restaurant identification", Details: []string{err.Error()}})
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid item uuid"})
		return
	}

	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid request body", Details: []string{err.Error()}})
		return
	}

	if err := h.menuService.UpdateStock(c.Request.Context(), restaurantID, itemID, *req.StockQuantity); err != nil {
		handler.RespondError(c, err)
		return
	}

	handler.RespondSuccess(c, "item stock successfully updated")
}

// SyncMenu полностью обновляет категории и блюда ресторана.
func (h *Handler) SyncMenu(c *gin.Context) {
	restaurantID, err := getPartnerRestaurantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid restaurant identification", Details: []string{err.Error()}})
		return
	}

	var req SyncMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid request body", Details: []string{err.Error()}})
		return
	}

	categories := make([]domain.MenuCategory, len(req.Categories))
	for i, cat := range req.Categories {
		categoryID := uuid.New()
		items := make([]domain.MenuItem, len(cat.Items))
		for j, it := range cat.Items {
			items[j] = domain.MenuItem{
				ID:            uuid.New(),
				RestaurantID:  restaurantID,
				CategoryID:    categoryID,
				Name:          it.Name,
				Description:   it.Description,
				Price:         it.Price,
				StockQuantity: it.StockQuantity,
				IsAvailable:   it.IsAvailable,
			}
		}

		categories[i] = domain.MenuCategory{
			ID:           categoryID,
			RestaurantID: restaurantID,
			Name:         cat.Name,
			SortOrder:    cat.SortOrder,
			Items:        items,
		}
	}

	if err := h.menuService.UpdateMenu(c.Request.Context(), restaurantID, categories); err != nil {
		handler.RespondError(c, err)
		return
	}

	handler.RespondSuccess(c, "menu synchronized successfully")
}

// ListOrders возвращает историю и активные заказы заведения с фильтрацией по статусу.
func (h *Handler) ListOrders(c *gin.Context) {
	restaurantID, err := getPartnerRestaurantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid restaurant identification", Details: []string{err.Error()}})
		return
	}

	var statusFilter *domain.OrderStatus
	if statusStr := c.Query("status"); statusStr != "" {
		st := domain.OrderStatus(statusStr)
		statusFilter = &st
	}

	orders, err := h.orderService.ListRestaurantOrders(c.Request.Context(), restaurantID, statusFilter)
	if err != nil {
		handler.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus переводит заказ на следующий шаг жизненного цикла по правилам стейт-машины.
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid order uuid"})
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, handler.ErrorResponse{Error: "invalid request body", Details: []string{err.Error()}})
		return
	}

	if err := h.orderService.UpdateStatus(c.Request.Context(), orderID, req.Status, req.Comment); err != nil {
		handler.RespondError(c, err)
		return
	}

	handler.RespondSuccess(c, "order status successfully updated")
}
