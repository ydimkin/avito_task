package service

import (
	"context"
	"testing"
	"time"
)

type mockCoreClient struct {
	lastOrderID   string
	lastStatus    string
	lastComment   *string
	stopListRest  string
	stopListItem  string
	stopListAvail bool
}

func (m *mockCoreClient) UpdateOrderStatus(ctx context.Context, orderID string, status string, comment *string) error {
	m.lastOrderID = orderID
	m.lastStatus = status
	m.lastComment = comment
	return nil
}

func (m *mockCoreClient) SetItemAvailability(ctx context.Context, restaurantID, itemID string, isAvailable bool) error {
	m.stopListRest = restaurantID
	m.stopListItem = itemID
	m.stopListAvail = isAvailable
	return nil
}

func TestKitchenSimulator_HandleIncomingOrderAndGetOrders(t *testing.T) {
	client := &mockCoreClient{}
	sim := NewKitchenSimulator(client, false) // autoCook = false for deterministic test

	order := ReceivedOrder{
		ID:              "order-123",
		RestaurantID:    "rest-1",
		CustomerName:    "Тестовый Клиент",
		CustomerPhone:   "+79998887766",
		DeliveryAddress: "Москва",
		Status:          "CREATED",
		TotalPrice:      500.0,
	}

	sim.HandleIncomingOrder(context.Background(), order)

	orders := sim.GetOrders()
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if orders[0].ID != "order-123" {
		t.Fatalf("expected order ID 'order-123', got %s", orders[0].ID)
	}
	if orders[0].ReceivedAt.IsZero() {
		t.Fatal("expected ReceivedAt to be set")
	}
}

func TestKitchenSimulator_TriggerStopList(t *testing.T) {
	client := &mockCoreClient{}
	sim := NewKitchenSimulator(client, false)

	err := sim.TriggerStopList(context.Background(), "rest-1", "item-99", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.stopListRest != "rest-1" || client.stopListItem != "item-99" || client.stopListAvail != false {
		t.Fatalf("stop list trigger parameters mismatch: %+v", client)
	}
}

func TestKitchenSimulator_ConcurrentAccess(t *testing.T) {
	client := &mockCoreClient{}
	sim := NewKitchenSimulator(client, false)

	for i := 0; i < 20; i++ {
		go func(idx int) {
			sim.HandleIncomingOrder(context.Background(), ReceivedOrder{
				ID: "concurrent-order",
			})
		}(i)
		go func() {
			_ = sim.GetOrders()
		}()
	}

	time.Sleep(50 * time.Millisecond)
	orders := sim.GetOrders()
	if len(orders) == 0 {
		t.Fatal("expected non-empty orders list")
	}
}
