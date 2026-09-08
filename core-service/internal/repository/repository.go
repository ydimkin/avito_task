package repository

import (
	"context"

	"core-service/internal/domain"
	"github.com/google/uuid"
)

// RestaurantRepository определяет контракт доступа к данным заведений.
type RestaurantRepository interface {
	List(ctx context.Context, search string, limit, offset int) ([]domain.Restaurant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error)
	Create(ctx context.Context, r *domain.Restaurant) error
}

// MenuRepository определяет контракт доступа к меню заведений.
type MenuRepository interface {
	GetMenuByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]domain.MenuCategory, error)
	GetItemsByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.MenuItem, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error)
	SetItemAvailability(ctx context.Context, restaurantID, itemID uuid.UUID, isAvailable bool) error
	UpdateStock(ctx context.Context, restaurantID, itemID uuid.UUID, stock int) error
	UpsertMenu(ctx context.Context, restaurantID uuid.UUID, categories []domain.MenuCategory) error
}

// OrderRepository определяет контракт доступа к данным заказов и их истории.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, newStatus domain.OrderStatus, comment *string) error
	ListOrdersByRestaurant(ctx context.Context, restaurantID uuid.UUID, status *domain.OrderStatus) ([]domain.Order, error)
}
