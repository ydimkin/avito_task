package domain

import (
	"testing"
)

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     OrderStatus
		to       OrderStatus
		expected bool
	}{
		// Допустимые переходы
		{"CREATED to ACCEPTED", OrderStatusCreated, OrderStatusAccepted, true},
		{"CREATED to CANCELLED", OrderStatusCreated, OrderStatusCancelled, true},
		{"ACCEPTED to COOKING", OrderStatusAccepted, OrderStatusCooking, true},
		{"ACCEPTED to CANCELLED", OrderStatusAccepted, OrderStatusCancelled, true},
		{"COOKING to READY_FOR_DELIVERY", OrderStatusCooking, OrderStatusReadyForDelivery, true},
		{"COOKING to CANCELLED", OrderStatusCooking, OrderStatusCancelled, true},
		{"READY_FOR_DELIVERY to IN_DELIVERY", OrderStatusReadyForDelivery, OrderStatusInDelivery, true},
		{"READY_FOR_DELIVERY to CANCELLED", OrderStatusReadyForDelivery, OrderStatusCancelled, true},
		{"IN_DELIVERY to DELIVERED", OrderStatusInDelivery, OrderStatusDelivered, true},
		{"IN_DELIVERY to CANCELLED", OrderStatusInDelivery, OrderStatusCancelled, true},

		// Недопустимые переходы
		{"CREATED to COOKING (skip stage)", OrderStatusCreated, OrderStatusCooking, false},
		{"CREATED to DELIVERED (direct finish)", OrderStatusCreated, OrderStatusDelivered, false},
		{"ACCEPTED to DELIVERED", OrderStatusAccepted, OrderStatusDelivered, false},
		{"COOKING to ACCEPTED (reverse)", OrderStatusCooking, OrderStatusAccepted, false},
		{"DELIVERED to CANCELLED (terminal)", OrderStatusDelivered, OrderStatusCancelled, false},
		{"DELIVERED to COOKING (terminal)", OrderStatusDelivered, OrderStatusCooking, false},
		{"CANCELLED to ACCEPTED (terminal)", OrderStatusCancelled, OrderStatusAccepted, false},
		{"CANCELLED to CREATED (terminal)", OrderStatusCancelled, OrderStatusCreated, false},
		{"Unknown from status", OrderStatus("UNKNOWN"), OrderStatusCreated, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidTransition(tc.from, tc.to)
			if result != tc.expected {
				t.Errorf("IsValidTransition(%s, %s) = %v, expected %v", tc.from, tc.to, result, tc.expected)
			}
		})
	}
}

func TestCanCustomerCancel(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected bool
	}{
		{OrderStatusCreated, true},
		{OrderStatusAccepted, true},
		{OrderStatusCooking, false},
		{OrderStatusReadyForDelivery, false},
		{OrderStatusInDelivery, false},
		{OrderStatusDelivered, false},
		{OrderStatusCancelled, false},
		{OrderStatus("UNKNOWN"), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			result := CanCustomerCancel(tc.status)
			if result != tc.expected {
				t.Errorf("CanCustomerCancel(%s) = %v, expected %v", tc.status, result, tc.expected)
			}
		})
	}
}

func TestIsTerminalStatus(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected bool
	}{
		{OrderStatusDelivered, true},
		{OrderStatusCancelled, true},
		{OrderStatusCreated, false},
		{OrderStatusAccepted, false},
		{OrderStatusCooking, false},
		{OrderStatusReadyForDelivery, false},
		{OrderStatusInDelivery, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			result := IsTerminalStatus(tc.status)
			if result != tc.expected {
				t.Errorf("IsTerminalStatus(%s) = %v, expected %v", tc.status, result, tc.expected)
			}
		})
	}
}
