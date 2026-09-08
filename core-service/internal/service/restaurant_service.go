package service

import (
	"context"

	"core-service/internal/domain"
	"core-service/internal/repository"
	"github.com/google/uuid"
)

type restaurantService struct {
	repo repository.RestaurantRepository
}

// NewRestaurantService создает экземпляр сервиса заведений.
func NewRestaurantService(repo repository.RestaurantRepository) RestaurantService {
	return &restaurantService{repo: repo}
}

func (s *restaurantService) List(ctx context.Context, search string, limit, offset int) ([]domain.Restaurant, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	restaurants, err := s.repo.List(ctx, search, limit, offset)
	if err != nil {
		return nil, err
	}
	return restaurants, nil
}

func (s *restaurantService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	restaurant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return restaurant, nil
}
