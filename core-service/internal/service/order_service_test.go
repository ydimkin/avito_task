package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"core-service/internal/domain"

	"github.com/google/uuid"
)

// Mock OrderRepository
type mockOrderRepo struct {
	createdOrder *domain.Order
	orders       map[uuid.UUID]*domain.Order
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: make(map[uuid.UUID]*domain.Order)}
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *domain.Order) error {
	m.createdOrder = order
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepo) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	if o, ok := m.orders[id]; ok {
		return o, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockOrderRepo) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, newStatus domain.OrderStatus, comment *string) error {
	if o, ok := m.orders[orderID]; ok {
		o.Status = newStatus
		return nil
	}
	return domain.ErrNotFound
}

func (m *mockOrderRepo) CancelOrder(ctx context.Context, orderID uuid.UUID, comment *string) error {
	o, ok := m.orders[orderID]
	if !ok {
		return domain.ErrNotFound
	}
	if !domain.CanCustomerCancel(o.Status) {
		return domain.ErrCannotCancelOrder
	}
	o.Status = domain.OrderStatusCancelled
	return nil
}

func (m *mockOrderRepo) ListOrdersByRestaurant(ctx context.Context, restaurantID uuid.UUID, status *domain.OrderStatus) ([]domain.Order, error) {
	var list []domain.Order
	for _, o := range m.orders {
		if o.RestaurantID == restaurantID {
			if status == nil || o.Status == *status {
				list = append(list, *o)
			}
		}
	}
	return list, nil
}

// Mock MenuRepository
type mockMenuRepo struct {
	items map[uuid.UUID]domain.MenuItem
}

func (m *mockMenuRepo) GetMenuByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]domain.MenuCategory, error) {
	return nil, nil
}

func (m *mockMenuRepo) GetItemsByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.MenuItem, error) {
	var res []domain.MenuItem
	for _, id := range ids {
		if item, ok := m.items[id]; ok {
			res = append(res, item)
		}
	}
	return res, nil
}

func (m *mockMenuRepo) GetItemByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	if item, ok := m.items[id]; ok {
		return &item, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockMenuRepo) SetItemAvailability(ctx context.Context, restaurantID, itemID uuid.UUID, isAvailable bool) error {
	return nil
}

func (m *mockMenuRepo) UpdateStock(ctx context.Context, restaurantID, itemID uuid.UUID, stock int) error {
	return nil
}

func (m *mockMenuRepo) UpsertMenu(ctx context.Context, restaurantID uuid.UUID, categories []domain.MenuCategory) error {
	return nil
}

// Mock RestaurantRepository
type mockRestaurantRepo struct {
	restaurants map[uuid.UUID]domain.Restaurant
}

func (m *mockRestaurantRepo) List(ctx context.Context, search string, limit, offset int) ([]domain.Restaurant, error) {
	return nil, nil
}

func (m *mockRestaurantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	if r, ok := m.restaurants[id]; ok {
		return &r, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockRestaurantRepo) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	return nil
}

// Mock WebhookNotifier
type mockWebhookNotifier struct {
	notified bool
}

func (m *mockWebhookNotifier) NotifyNewOrder(ctx context.Context, webhookURL string, order *domain.Order) error {
	m.notified = true
	return nil
}

func TestOrderService_CreateOrder(t *testing.T) {
	ctx := context.Background()
	restID := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()

	activeRestaurant := domain.Restaurant{
		ID:         restID,
		Name:       "Test Rest",
		IsActive:   true,
		WebhookURL: "http://test/webhook",
	}
	inactiveRestID := uuid.New()
	inactiveRestaurant := domain.Restaurant{
		ID:       inactiveRestID,
		Name:     "Closed Rest",
		IsActive: false,
	}

	items := map[uuid.UUID]domain.MenuItem{
		itemID1: {
			ID:            itemID1,
			RestaurantID:  restID,
			Name:          "Бургер",
			Price:         300.0,
			StockQuantity: 10,
			IsAvailable:   true,
		},
		itemID2: {
			ID:            itemID2,
			RestaurantID:  restID,
			Name:          "Кола",
			Price:         100.0,
			StockQuantity: 5,
			IsAvailable:   true,
		},
	}

	t.Run("empty items returns ErrEmptyOrder", func(t *testing.T) {
		svc := NewOrderService(newMockOrderRepo(), &mockMenuRepo{items: items}, &mockRestaurantRepo{}, &mockWebhookNotifier{})
		_, err := svc.CreateOrder(ctx, CreateOrderDTO{
			RestaurantID: restID,
			Items:        nil,
		})
		if !errors.Is(err, domain.ErrEmptyOrder) {
			t.Fatalf("expected ErrEmptyOrder, got %v", err)
		}
	})

	t.Run("inactive restaurant returns ErrRestaurantInactive", func(t *testing.T) {
		restRepo := &mockRestaurantRepo{
			restaurants: map[uuid.UUID]domain.Restaurant{inactiveRestID: inactiveRestaurant},
		}
		svc := NewOrderService(newMockOrderRepo(), &mockMenuRepo{items: items}, restRepo, &mockWebhookNotifier{})
		_, err := svc.CreateOrder(ctx, CreateOrderDTO{
			RestaurantID: inactiveRestID,
			Items:        []OrderItemInput{{MenuItemID: itemID1, Quantity: 1}},
		})
		if !errors.Is(err, domain.ErrRestaurantInactive) {
			t.Fatalf("expected ErrRestaurantInactive, got %v", err)
		}
	})

	t.Run("invalid quantity <= 0 returns ErrInvalidInput", func(t *testing.T) {
		restRepo := &mockRestaurantRepo{
			restaurants: map[uuid.UUID]domain.Restaurant{restID: activeRestaurant},
		}
		svc := NewOrderService(newMockOrderRepo(), &mockMenuRepo{items: items}, restRepo, &mockWebhookNotifier{})
		_, err := svc.CreateOrder(ctx, CreateOrderDTO{
			RestaurantID: restID,
			Items:        []OrderItemInput{{MenuItemID: itemID1, Quantity: 0}},
		})
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("item in stop-list returns ErrItemInStopList", func(t *testing.T) {
		stopItemID := uuid.New()
		menuItems := map[uuid.UUID]domain.MenuItem{
			stopItemID: {
				ID:            stopItemID,
				RestaurantID:  restID,
				Name:          "Стоп-Блюдо",
				Price:         200,
				StockQuantity: 5,
				IsAvailable:   false, // Стоп-лист
			},
		}
		restRepo := &mockRestaurantRepo{
			restaurants: map[uuid.UUID]domain.Restaurant{restID: activeRestaurant},
		}
		svc := NewOrderService(newMockOrderRepo(), &mockMenuRepo{items: menuItems}, restRepo, &mockWebhookNotifier{})
		_, err := svc.CreateOrder(ctx, CreateOrderDTO{
			RestaurantID: restID,
			Items:        []OrderItemInput{{MenuItemID: stopItemID, Quantity: 1}},
		})
		if !errors.Is(err, domain.ErrItemInStopList) {
			t.Fatalf("expected ErrItemInStopList, got %v", err)
		}
	})

	t.Run("out of stock returns ErrOutOfStock", func(t *testing.T) {
		restRepo := &mockRestaurantRepo{
			restaurants: map[uuid.UUID]domain.Restaurant{restID: activeRestaurant},
		}
		svc := NewOrderService(newMockOrderRepo(), &mockMenuRepo{items: items}, restRepo, &mockWebhookNotifier{})
		_, err := svc.CreateOrder(ctx, CreateOrderDTO{
			RestaurantID: restID,
			Items:        []OrderItemInput{{MenuItemID: itemID2, Quantity: 10}}, // 10 > 5
		})
		if !errors.Is(err, domain.ErrOutOfStock) {
			t.Fatalf("expected ErrOutOfStock, got %v", err)
		}
	})

	t.Run("successful order creation calculates total price correctly", func(t *testing.T) {
		orderRepo := newMockOrderRepo()
		restRepo := &mockRestaurantRepo{
			restaurants: map[uuid.UUID]domain.Restaurant{restID: activeRestaurant},
		}
		svc := NewOrderService(orderRepo, &mockMenuRepo{items: items}, restRepo, &mockWebhookNotifier{})

		order, err := svc.CreateOrder(ctx, CreateOrderDTO{
			RestaurantID:    restID,
			CustomerName:    "Иван",
			CustomerPhone:   "+79991112233",
			DeliveryAddress: "Москва",
			Items: []OrderItemInput{
				{MenuItemID: itemID1, Quantity: 2}, // 2 * 300 = 600
				{MenuItemID: itemID2, Quantity: 3}, // 3 * 100 = 300
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Status != domain.OrderStatusCreated {
			t.Fatalf("expected CREATED status, got %s", order.Status)
		}
		if order.TotalPrice != 900.0 {
			t.Fatalf("expected total price 900.0, got %f", order.TotalPrice)
		}
		if len(order.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(order.Items))
		}
		if orderRepo.createdOrder == nil {
			t.Fatal("expected order to be persisted in repo")
		}
	})
}

func TestOrderService_CancelOrder(t *testing.T) {
	ctx := context.Background()
	orderRepo := newMockOrderRepo()
	svc := NewOrderService(orderRepo, &mockMenuRepo{}, &mockRestaurantRepo{}, nil)

	createdID := uuid.New()
	orderRepo.orders[createdID] = &domain.Order{
		ID:        createdID,
		Status:    domain.OrderStatusCreated,
		CreatedAt: time.Now(),
	}

	cookingID := uuid.New()
	orderRepo.orders[cookingID] = &domain.Order{
		ID:        cookingID,
		Status:    domain.OrderStatusCooking,
		CreatedAt: time.Now(),
	}

	t.Run("cancel created order succeeds", func(t *testing.T) {
		err := svc.CancelOrder(ctx, createdID, "клиент передумал")
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if orderRepo.orders[createdID].Status != domain.OrderStatusCancelled {
			t.Fatalf("expected status CANCELLED, got %s", orderRepo.orders[createdID].Status)
		}
	})

	t.Run("cancel cooking order fails with ErrCannotCancelOrder", func(t *testing.T) {
		err := svc.CancelOrder(ctx, cookingID, "слишком долго")
		if !errors.Is(err, domain.ErrCannotCancelOrder) {
			t.Fatalf("expected ErrCannotCancelOrder, got %v", err)
		}
	})
}

func TestOrderService_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	orderRepo := newMockOrderRepo()
	svc := NewOrderService(orderRepo, &mockMenuRepo{}, &mockRestaurantRepo{}, nil)

	orderID := uuid.New()
	orderRepo.orders[orderID] = &domain.Order{
		ID:        orderID,
		Status:    domain.OrderStatusCreated,
		CreatedAt: time.Now(),
	}

	t.Run("valid transition CREATED to ACCEPTED succeeds", func(t *testing.T) {
		err := svc.UpdateStatus(ctx, orderID, domain.OrderStatusAccepted, nil)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if orderRepo.orders[orderID].Status != domain.OrderStatusAccepted {
			t.Fatalf("expected status ACCEPTED, got %s", orderRepo.orders[orderID].Status)
		}
	})

	t.Run("invalid transition ACCEPTED to DELIVERED fails with ErrInvalidStatusTransition", func(t *testing.T) {
		err := svc.UpdateStatus(ctx, orderID, domain.OrderStatusDelivered, nil)
		if !errors.Is(err, domain.ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}
