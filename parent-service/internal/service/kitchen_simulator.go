package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"parent-service/internal/client"
)

// ReceivedOrder структура для хранения полученного через Webhook заказа.
type ReceivedOrder struct {
	ID              string      `json:"id"`
	RestaurantID    string      `json:"restaurant_id"`
	CustomerName    string      `json:"customer_name"`
	CustomerPhone   string      `json:"customer_phone"`
	DeliveryAddress string      `json:"delivery_address"`
	Status          string      `json:"status"`
	TotalPrice      float64     `json:"total_price"`
	Items           interface{} `json:"items"`
	ReceivedAt      time.Time   `json:"received_at"`
}

// KitchenSimulator интерфейс эмулятора работы кухни ресторана.
type KitchenSimulator interface {
	HandleIncomingOrder(ctx context.Context, order ReceivedOrder)
	GetOrders() []ReceivedOrder
	TriggerStopList(ctx context.Context, restaurantID, itemID string, isAvailable bool) error
}

type kitchenSimulator struct {
	client   client.CoreClient
	ordersMu sync.RWMutex
	orders   []ReceivedOrder
	autoCook bool
}

// NewKitchenSimulator создает экземпляр симулятора кухни.
func NewKitchenSimulator(client client.CoreClient, autoCook bool) KitchenSimulator {
	return &kitchenSimulator{
		client:   client,
		orders:   make([]ReceivedOrder, 0),
		autoCook: autoCook,
	}
}

func (s *kitchenSimulator) HandleIncomingOrder(ctx context.Context, order ReceivedOrder) {
	order.ReceivedAt = time.Now()
	s.ordersMu.Lock()
	s.orders = append(s.orders, order)
	s.ordersMu.Unlock()

	slog.Info("received incoming order webhook",
		slog.String("order_id", order.ID),
		slog.Float64("total_price", order.TotalPrice),
	)

	if s.autoCook {
		go s.simulateCookingProcess(order.ID)
	}
}

func (s *kitchenSimulator) simulateCookingProcess(orderID string) {
	steps := []struct {
		status  string
		comment string
		delay   time.Duration
	}{
		{
			status:  "ACCEPTED",
			comment: "Ресторан подтвердил получение и принял заказ",
			delay:   2 * time.Second,
		},
		{
			status:  "COOKING",
			comment: "Шеф-повар начал приготовление блюд",
			delay:   3 * time.Second,
		},
		{
			status:  "READY_FOR_DELIVERY",
			comment: "Заказ упакован и ожидает курьера",
			delay:   3 * time.Second,
		},
		{
			status:  "IN_DELIVERY",
			comment: "Курьер забрал заказ и направляется к клиенту",
			delay:   3 * time.Second,
		},
		{
			status:  "DELIVERED",
			comment: "Заказ успешно вручен клиенту",
			delay:   3 * time.Second,
		},
	}

	for _, step := range steps {
		time.Sleep(step.delay)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := s.client.UpdateOrderStatus(ctx, orderID, step.status, &step.comment)
		cancel()
		if err != nil {
			slog.Error("failed to update order status",
				slog.String("order_id", orderID),
				slog.String("status", step.status),
				slog.Any("error", err),
			)
			return
		}
		slog.Info("order status advanced by kitchen",
			slog.String("order_id", orderID),
			slog.String("status", step.status),
			slog.String("comment", step.comment),
		)
	}
}

func (s *kitchenSimulator) GetOrders() []ReceivedOrder {
	s.ordersMu.RLock()
	defer s.ordersMu.RUnlock()

	result := make([]ReceivedOrder, len(s.orders))
	copy(result, s.orders)
	return result
}

func (s *kitchenSimulator) TriggerStopList(ctx context.Context, restaurantID, itemID string, isAvailable bool) error {
	return s.client.SetItemAvailability(ctx, restaurantID, itemID, isAvailable)
}
