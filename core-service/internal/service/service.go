package service

import (
	"context"

	"core-service/internal/domain"
	"github.com/google/uuid"
)

// OrderItemInput описывает элемент заказа при его создании.
type OrderItemInput struct {
	MenuItemID uuid.UUID `json:"menu_item_id"`
	Quantity   int       `json:"quantity"`
}

// CreateOrderDTO содержит данные, необходимые для оформления заказа.
type CreateOrderDTO struct {
	UserID          *uuid.UUID       `json:"user_id,omitempty"`
	RestaurantID    uuid.UUID        `json:"restaurant_id"`
	CustomerName    string           `json:"customer_name"`
	CustomerPhone   string           `json:"customer_phone"`
	DeliveryAddress string           `json:"delivery_address"`
	Items           []OrderItemInput `json:"items"`
}

// RestaurantService контракт сервиса работы с заведениями.
type RestaurantService interface {
	List(ctx context.Context, search string, limit, offset int) ([]domain.Restaurant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error)
}

// MenuService контракт сервиса управления меню и стоп-листами.
type MenuService interface {
	GetMenu(ctx context.Context, restaurantID uuid.UUID) ([]domain.MenuCategory, error)
	SetAvailability(ctx context.Context, restaurantID, itemID uuid.UUID, isAvailable bool) error
	UpdateStock(ctx context.Context, restaurantID, itemID uuid.UUID, stock int) error
	UpdateMenu(ctx context.Context, restaurantID uuid.UUID, categories []domain.MenuCategory) error
}

// OrderService контракт сервиса заказов.
type OrderService interface {
	CreateOrder(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	CancelOrder(ctx context.Context, id uuid.UUID, reason string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, newStatus domain.OrderStatus, comment *string) error
	ListRestaurantOrders(ctx context.Context, restaurantID uuid.UUID, status *domain.OrderStatus) ([]domain.Order, error)
}

// WebhookNotifier контракт отправщика вебхуков в заведения.
type WebhookNotifier interface {
	NotifyNewOrder(ctx context.Context, webhookURL string, order *domain.Order) error
}
