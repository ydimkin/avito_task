package repository

import (
	"context"
	"errors"

	"core-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postgresRestaurantRepository struct {
	db *gorm.DB
}

// NewPostgresRestaurantRepository создает новый экземпляр репозитория ресторанов.
func NewPostgresRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &postgresRestaurantRepository{db: db}
}

func (r *postgresRestaurantRepository) List(ctx context.Context, search string, limit, offset int) ([]domain.Restaurant, error) {
	var restaurants []domain.Restaurant
	query := r.db.WithContext(ctx).Where("is_active = ?", true)

	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&restaurants).Error; err != nil {
		return nil, err
	}

	return restaurants, nil
}

func (r *postgresRestaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	var restaurant domain.Restaurant
	err := r.db.WithContext(ctx).
		Preload("Categories.Items").
		First(&restaurant, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &restaurant, nil
}

func (r *postgresRestaurantRepository) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	if err := r.db.WithContext(ctx).Create(restaurant).Error; err != nil {
		return err
	}
	return nil
}
