package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus представляет допустимые статусы заказа
type OrderStatus string

// Статусы заказа в системе
const (
	OrderStatusCreated          OrderStatus = "CREATED"
	OrderStatusAccepted         OrderStatus = "ACCEPTED"
	OrderStatusCooking          OrderStatus = "COOKING"
	OrderStatusReadyForDelivery OrderStatus = "READY_FOR_DELIVERY"
	OrderStatusInDelivery       OrderStatus = "IN_DELIVERY"
	OrderStatusDelivered        OrderStatus = "DELIVERED"
	OrderStatusCancelled        OrderStatus = "CANCELLED"
)

// Restaurant представляет модель заведения-партнера
type Restaurant struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(255);not null"`
	Description string         `json:"description" gorm:"type:text"`
	Address     string         `json:"address" gorm:"type:text;not null"`
	WebhookURL  string         `json:"webhook_url" gorm:"type:varchar(512)"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index"`
	Categories  []MenuCategory `json:"categories,omitempty" gorm:"foreignKey:RestaurantID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// MenuCategory представляет категорию меню заведения
type MenuCategory struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RestaurantID uuid.UUID  `json:"restaurant_id" gorm:"type:uuid;not null;index"`
	Name         string     `json:"name" gorm:"type:varchar(255);not null"`
	SortOrder    int        `json:"sort_order" gorm:"default:0"`
	Items        []MenuItem `json:"items,omitempty" gorm:"foreignKey:CategoryID"`
	CreatedAt    time.Time  `json:"created_at"`
}

// MenuItem представляет позицию меню
type MenuItem struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RestaurantID    uuid.UUID `json:"restaurant_id" gorm:"type:uuid;not null;index"`
	CategoryID      uuid.UUID `json:"category_id" gorm:"type:uuid;not null;index"`
	Name            string    `json:"name" gorm:"type:varchar(255);not null"`
	Description     string    `json:"description" gorm:"type:text"`
	Price           float64   `json:"price" gorm:"type:numeric(10,2);not null"`
	StockQuantity   int       `json:"stock_quantity" gorm:"not null;default:0"`
	IsAvailable     bool      `json:"is_available" gorm:"default:true;index"`
	AutoUnavailable bool      `json:"-" gorm:"not null;default:false"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Order представляет заказ клиента
type Order struct {
	ID              uuid.UUID            `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          *uuid.UUID           `json:"user_id,omitempty" gorm:"type:uuid;index"`
	RestaurantID    uuid.UUID            `json:"restaurant_id" gorm:"type:uuid;not null;index"`
	CustomerName    string               `json:"customer_name" gorm:"type:varchar(255);not null"`
	CustomerPhone   string               `json:"customer_phone" gorm:"type:varchar(50);not null"`
	DeliveryAddress string               `json:"delivery_address" gorm:"type:text;not null"`
	Status          OrderStatus          `json:"status" gorm:"type:varchar(50);not null;index"`
	TotalPrice      float64              `json:"total_price" gorm:"type:numeric(10,2);not null"`
	CancelReason    *string              `json:"cancel_reason,omitempty" gorm:"type:text"`
	Items           []OrderItem          `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	StatusHistory   []OrderStatusHistory `json:"status_history,omitempty" gorm:"foreignKey:OrderID"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// OrderItem представляет позицию в заказе с фиксацией цены и наименования на момент покупки
type OrderItem struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrderID       uuid.UUID `json:"order_id" gorm:"type:uuid;not null;index"`
	MenuItemID    uuid.UUID `json:"menu_item_id" gorm:"type:uuid;not null"`
	NameAtMoment  string    `json:"name_at_moment" gorm:"type:varchar(255);not null"`
	PriceAtMoment float64   `json:"price_at_moment" gorm:"type:numeric(10,2);not null"`
	Quantity      int       `json:"quantity" gorm:"not null"`
	Subtotal      float64   `json:"subtotal" gorm:"type:numeric(10,2);not null"`
}

// OrderStatusHistory сохраняет историю изменения статусов заказа
type OrderStatusHistory struct {
	ID        uuid.UUID   `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrderID   uuid.UUID   `json:"order_id" gorm:"type:uuid;not null;index"`
	Status    OrderStatus `json:"status" gorm:"type:varchar(50);not null"`
	Comment   *string     `json:"comment,omitempty" gorm:"type:text"`
	CreatedAt time.Time   `json:"created_at"`
}

// TableName реализует интерфейс schema.Tabler для GORM
func (OrderStatusHistory) TableName() string {
	return "order_status_history"
}
