package domain

// validTransitions задает карту допустимых переходов стейт-машины заказа.
var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusCreated: {
		OrderStatusAccepted,
		OrderStatusCancelled,
	},
	OrderStatusAccepted: {
		OrderStatusCooking,
		OrderStatusCancelled,
	},
	OrderStatusCooking: {
		OrderStatusReadyForDelivery,
		OrderStatusCancelled,
	},
	OrderStatusReadyForDelivery: {
		OrderStatusInDelivery,
		OrderStatusCancelled,
	},
	OrderStatusInDelivery: {
		OrderStatusDelivered,
		OrderStatusCancelled,
	},
	OrderStatusDelivered: {},
	OrderStatusCancelled: {},
}

// Проверяет, допустим ли переход между статусами заказа.
func IsValidTransition(from, to OrderStatus) bool {
	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

// Проверяет, имеет ли клиент право отменить заказ.
// Клиент может отменить заказ только до начала приготовления (в статусах CREATED и ACCEPTED).
func CanCustomerCancel(current OrderStatus) bool {
	return current == OrderStatusCreated || current == OrderStatusAccepted
}

// IsTerminalStatus проверяет, является ли статус конечным.
func IsTerminalStatus(status OrderStatus) bool {
	return status == OrderStatusDelivered || status == OrderStatusCancelled
}
