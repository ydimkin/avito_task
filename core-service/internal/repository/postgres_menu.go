package repository

import (
	"context"
	"errors"

	"core-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postgresMenuRepository struct {
	db *gorm.DB
}

// NewPostgresMenuRepository создает экземпляр репозитория меню на GORM.
func NewPostgresMenuRepository(db *gorm.DB) MenuRepository {
	return &postgresMenuRepository{db: db}
}

func (r *postgresMenuRepository) GetMenuByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]domain.MenuCategory, error) {
	var categories []domain.MenuCategory
	err := r.db.WithContext(ctx).
		Where("restaurant_id = ?", restaurantID).
		Order("sort_order ASC").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *postgresMenuRepository) GetItemsByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.MenuItem, error) {
	var items []domain.MenuItem
	if len(ids) == 0 {
		return items, nil
	}

	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *postgresMenuRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	var item domain.MenuItem
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *postgresMenuRepository) SetItemAvailability(ctx context.Context, restaurantID, itemID uuid.UUID, isAvailable bool) error {
	result := r.db.WithContext(ctx).
		Model(&domain.MenuItem{}).
		Where("id = ? AND restaurant_id = ?", itemID, restaurantID).
		Update("is_available", isAvailable)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *postgresMenuRepository) UpdateStock(ctx context.Context, restaurantID, itemID uuid.UUID, stock int) error {
	if stock < 0 {
		return domain.ErrInvalidInput
	}

	isAvailable := stock > 0
	result := r.db.WithContext(ctx).
		Model(&domain.MenuItem{}).
		Where("id = ? AND restaurant_id = ?", itemID, restaurantID).
		Updates(map[string]interface{}{
			"stock_quantity": stock,
			"is_available":   isAvailable,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *postgresMenuRepository) UpsertMenu(ctx context.Context, restaurantID uuid.UUID, categories []domain.MenuCategory) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Очищаем старые категории и позиции (каскадно удалятся позиции)
		if err := tx.Where("restaurant_id = ?", restaurantID).Delete(&domain.MenuCategory{}).Error; err != nil {
			return err
		}

		for i := range categories {
			categories[i].RestaurantID = restaurantID
			for j := range categories[i].Items {
				categories[i].Items[j].RestaurantID = restaurantID
			}
		}

		if len(categories) > 0 {
			if err := tx.Create(&categories).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
