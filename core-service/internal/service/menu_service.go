package service

import (
	"context"

	"core-service/internal/domain"
	"core-service/internal/repository"
	"github.com/google/uuid"
)

type menuService struct {
	menuRepo       repository.MenuRepository
	restaurantRepo repository.RestaurantRepository
}

// NewMenuService создает экземпляр сервиса управления меню.
func NewMenuService(menuRepo repository.MenuRepository, restaurantRepo repository.RestaurantRepository) MenuService {
	return &menuService{
		menuRepo:       menuRepo,
		restaurantRepo: restaurantRepo,
	}
}

func (s *menuService) GetMenu(ctx context.Context, restaurantID uuid.UUID) ([]domain.MenuCategory, error) {
	// Проверяем существование ресторана
	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	if !restaurant.IsActive {
		return nil, domain.ErrRestaurantInactive
	}

	return s.menuRepo.GetMenuByRestaurantID(ctx, restaurantID)
}

func (s *menuService) SetAvailability(ctx context.Context, restaurantID, itemID uuid.UUID, isAvailable bool) error {
	return s.menuRepo.SetItemAvailability(ctx, restaurantID, itemID, isAvailable)
}

func (s *menuService) UpdateStock(ctx context.Context, restaurantID, itemID uuid.UUID, stock int) error {
	return s.menuRepo.UpdateStock(ctx, restaurantID, itemID, stock)
}

func (s *menuService) UpdateMenu(ctx context.Context, restaurantID uuid.UUID, categories []domain.MenuCategory) error {
	// Проверяем ресторан
	if _, err := s.restaurantRepo.GetByID(ctx, restaurantID); err != nil {
		return err
	}
	return s.menuRepo.UpsertMenu(ctx, restaurantID, categories)
}
