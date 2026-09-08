package repository

import (
	"context"
	"errors"
	"time"

	"core-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type postgresOrderRepository struct {
	db *gorm.DB
}

// NewPostgresOrderRepository создает экземпляр репозитория заказов на GORM.
func NewPostgresOrderRepository(db *gorm.DB) OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// SELECT FOR UPDATE
		for _, it := range order.Items {
			var item domain.MenuItem
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&item, "id = ?", it.MenuItemID).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.ErrNotFound
				}
				return err
			}

			if !item.IsAvailable {
				return domain.ErrItemInStopList
			}
			if item.StockQuantity < it.Quantity {
				return domain.ErrOutOfStock
			}

			newStock := item.StockQuantity - it.Quantity
			updates := map[string]interface{}{
				"stock_quantity": newStock,
			}
			if newStock == 0 {
				updates["is_available"] = false
				updates["auto_unavailable"] = true
			}

			if err := tx.Model(&item).Updates(updates).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		historyEntry := domain.OrderStatusHistory{
			ID:        uuid.New(),
			OrderID:   order.ID,
			Status:    order.Status,
			CreatedAt: time.Now(),
		}
		comment := "Заказ оформлен клиентом"
		historyEntry.Comment = &comment

		if err := tx.Create(&historyEntry).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *postgresOrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("StatusHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&order, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (r *postgresOrderRepository) UpdateOrderStatus(
	ctx context.Context,
	orderID uuid.UUID,
	newStatus domain.OrderStatus,
	comment *string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := r.lockOrder(tx, orderID)
		if err != nil {
			return err
		}
		if !domain.IsValidTransition(order.Status, newStatus) {
			return domain.ErrInvalidStatusTransition
		}

		if newStatus == domain.OrderStatusCancelled {
			if err := r.restoreOrderStock(tx, order); err != nil {
				return err
			}
		}

		// Обновляем статус заказа после проверки состояния под блокировкой строки.
		updates := map[string]interface{}{
			"status":     newStatus,
			"updated_at": time.Now(),
		}
		if newStatus == domain.OrderStatusCancelled && comment != nil {
			updates["cancel_reason"] = *comment
		}

		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return err
		}

		// 3. Фиксируем запись в истории статусов
		history := domain.OrderStatusHistory{
			ID:        uuid.New(),
			OrderID:   orderID,
			Status:    newStatus,
			Comment:   comment,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}

		return nil
	})
}

// CancelOrder атомарно проверяет окно отмены и меняет статус заказа.
// Проверка выполняется после SELECT FOR UPDATE, поэтому параллельный переход
// кухни в COOKING не может быть обойден устаревшим чтением в сервисе.
func (r *postgresOrderRepository) CancelOrder(
	ctx context.Context,
	orderID uuid.UUID,
	comment *string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := r.lockOrder(tx, orderID)
		if err != nil {
			return err
		}
		if !domain.CanCustomerCancel(order.Status) {
			return domain.ErrCannotCancelOrder
		}
		if err := r.restoreOrderStock(tx, order); err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":     domain.OrderStatusCancelled,
			"updated_at": time.Now(),
		}
		if comment != nil {
			updates["cancel_reason"] = *comment
		}
		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return err
		}

		history := domain.OrderStatusHistory{
			ID:        uuid.New(),
			OrderID:   orderID,
			Status:    domain.OrderStatusCancelled,
			Comment:   comment,
			CreatedAt: time.Now(),
		}
		return tx.Create(&history).Error
	})
}

func (r *postgresOrderRepository) lockOrder(tx *gorm.DB, orderID uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Items").First(&order, "id = ?", orderID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *postgresOrderRepository) restoreOrderStock(tx *gorm.DB, order *domain.Order) error {
	for _, item := range order.Items {
		if err := tx.Model(&domain.MenuItem{}).
			Where("id = ?", item.MenuItemID).
			Updates(map[string]interface{}{
				"stock_quantity":   gorm.Expr("stock_quantity + ?", item.Quantity),
				"is_available":     gorm.Expr("CASE WHEN auto_unavailable = true THEN true ELSE is_available END"),
				"auto_unavailable": false,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresOrderRepository) ListOrdersByRestaurant(
	ctx context.Context,
	restaurantID uuid.UUID,
	status *domain.OrderStatus,
) ([]domain.Order, error) {
	var orders []domain.Order
	query := r.db.WithContext(ctx).
		Where("restaurant_id = ?", restaurantID).
		Preload("Items").
		Preload("StatusHistory").
		Order("created_at DESC")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}
