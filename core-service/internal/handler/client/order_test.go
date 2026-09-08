package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"core-service/internal/domain"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockOrderService struct {
	createOrderFunc func(ctx context.Context, dto service.CreateOrderDTO) (*domain.Order, error)
	getOrderFunc    func(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	cancelOrderFunc func(ctx context.Context, id uuid.UUID, reason string) error
}

func (m *mockOrderService) CreateOrder(ctx context.Context, dto service.CreateOrderDTO) (*domain.Order, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, dto)
	}
	return nil, nil
}

func (m *mockOrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	if m.getOrderFunc != nil {
		return m.getOrderFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOrderService) CancelOrder(ctx context.Context, id uuid.UUID, reason string) error {
	if m.cancelOrderFunc != nil {
		return m.cancelOrderFunc(ctx, id, reason)
	}
	return nil
}

func (m *mockOrderService) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus domain.OrderStatus, comment *string) error {
	return nil
}

func (m *mockOrderService) ListRestaurantOrders(ctx context.Context, restaurantID uuid.UUID, status *domain.OrderStatus) ([]domain.Order, error) {
	return nil, nil
}

func setupOrderRouter(svc service.OrderService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewOrderHandler(svc)
	r.POST("/api/v1/orders", h.CreateOrder)
	r.GET("/api/v1/orders/:id", h.GetOrder)
	r.POST("/api/v1/orders/:id/cancel", h.CancelOrder)
	return r
}

func TestOrderHandler_CreateOrder(t *testing.T) {
	orderID := uuid.New()
	restID := uuid.New()
	itemID := uuid.New()

	mockSvc := &mockOrderService{
		createOrderFunc: func(ctx context.Context, dto service.CreateOrderDTO) (*domain.Order, error) {
			return &domain.Order{
				ID:           orderID,
				RestaurantID: restID,
				CustomerName: dto.CustomerName,
				Status:       domain.OrderStatusCreated,
				TotalPrice:   350.0,
			}, nil
		},
	}

	router := setupOrderRouter(mockSvc)

	body := CreateOrderRequest{
		RestaurantID:    restID,
		CustomerName:    "Иван",
		CustomerPhone:   "+79991234567",
		DeliveryAddress: "ул. Пушкина",
		Items: []CreateOrderItemRequest{
			{MenuItemID: itemID, Quantity: 1},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestOrderHandler_CancelOrder_ConflictWhenCooking(t *testing.T) {
	orderID := uuid.New()

	mockSvc := &mockOrderService{
		cancelOrderFunc: func(ctx context.Context, id uuid.UUID, reason string) error {
			return domain.ErrCannotCancelOrder
		},
	}

	router := setupOrderRouter(mockSvc)

	body := CancelOrderRequest{Reason: "клиент передумал"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/"+orderID.String()+"/cancel", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409 Conflict, got %d, body: %s", w.Code, w.Body.String())
	}
}
