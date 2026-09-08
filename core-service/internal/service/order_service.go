package service

import (
	"context"
	"fmt"
	"log/slog"

	"core-service/internal/domain"
	"core-service/internal/repository"

	"github.com/google/uuid"
)

type orderService struct {
	orderRepo      repository.OrderRepository
	menuRepo       repository.MenuRepository
	restaurantRepo repository.RestaurantRepository
	webhookService WebhookNotifier
}

// NewOrderService создает сервис обработки заказов.
func NewOrderService(
	orderRepo repository.OrderRepository,
	menuRepo repository.MenuRepository,
	restaurantRepo repository.RestaurantRepository,
	webhookService WebhookNotifier,
) OrderService {
	return &orderService{
		orderRepo:      orderRepo,
		menuRepo:       menuRepo,
		restaurantRepo: restaurantRepo,
		webhookService: webhookService,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error) {
	if len(dto.Items) == 0 {
		return nil, domain.ErrEmptyOrder
	}

	// Проверяем ресторан
	restaurant, err := s.restaurantRepo.GetByID(ctx, dto.RestaurantID)
	if err != nil {
		return nil, err
	}
	if !restaurant.IsActive {
		return nil, domain.ErrRestaurantInactive
	}

	// Собираем ID позиций и проверяем количество
	itemMap := make(map[uuid.UUID]int, len(dto.Items))
	itemIDs := make([]uuid.UUID, 0, len(dto.Items))
	for _, it := range dto.Items {
		if it.Quantity <= 0 {
			return nil, domain.ErrInvalidInput
		}
		itemMap[it.MenuItemID] = it.Quantity
		itemIDs = append(itemIDs, it.MenuItemID)
	}

	// Получаем актуальные позиции меню
	dbItems, err := s.menuRepo.GetItemsByIDs(ctx, itemIDs)
	if err != nil {
		return nil, err
	}
	if len(dbItems) != len(itemIDs) {
		return nil, domain.ErrNotFound
	}

	// Валидируем принадлежность ресторану, стоп-листы и фиксируем снимок цен
	orderID := uuid.New()
	var totalPrice float64
	orderItems := make([]domain.OrderItem, 0, len(dbItems))

	for _, dbItem := range dbItems {
		if dbItem.RestaurantID != dto.RestaurantID {
			return nil, domain.ErrInvalidInput
		}
		if !dbItem.IsAvailable {
			return nil, domain.ErrItemInStopList
		}

		qty := itemMap[dbItem.ID]
		if dbItem.StockQuantity < qty {
			return nil, domain.ErrOutOfStock
		}
		subtotal := dbItem.Price * float64(qty)
		totalPrice += subtotal

		orderItems = append(orderItems, domain.OrderItem{
			ID:            uuid.New(),
			OrderID:       orderID,
			MenuItemID:    dbItem.ID,
			NameAtMoment:  dbItem.Name,
			PriceAtMoment: dbItem.Price,
			Quantity:      qty,
			Subtotal:      subtotal,
		})
	}

	// Формируем сущность заказа
	order := &domain.Order{
		ID:              orderID,
		UserID:          dto.UserID,
		RestaurantID:    dto.RestaurantID,
		CustomerName:    dto.CustomerName,
		CustomerPhone:   dto.CustomerPhone,
		DeliveryAddress: dto.DeliveryAddress,
		Status:          domain.OrderStatusCreated,
		TotalPrice:      totalPrice,
		Items:           orderItems,
	}

	// Сохраняем в транзакции через репозиторий
	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	// Асинхронно отправляем вебхук ресторану
	if restaurant.WebhookURL != "" && s.webhookService != nil {
		go func(webhookURL string, orderCopy domain.Order) {
			bgCtx := context.Background()
			if notifyErr := s.webhookService.NotifyNewOrder(bgCtx, webhookURL, &orderCopy); notifyErr != nil {
				slog.WarnContext(bgCtx, "failed to notify restaurant webhook",
					slog.String("webhook_url", webhookURL),
					slog.String("order_id", orderCopy.ID.String()),
					slog.Any("error", notifyErr),
				)
			}
		}(restaurant.WebhookURL, *order)
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.orderRepo.GetOrderByID(ctx, id)
}

func (s *orderService) CancelOrder(ctx context.Context, id uuid.UUID, reason string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}

	if !domain.CanCustomerCancel(order.Status) {
		return domain.ErrCannotCancelOrder
	}

	comment := fmt.Sprintf("Отмена клиентом. Причина: %s", reason)
	return s.orderRepo.UpdateOrderStatus(ctx, id, domain.OrderStatusCancelled, &comment)
}

func (s *orderService) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus domain.OrderStatus, comment *string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}

	if !domain.IsValidTransition(order.Status, newStatus) {
		return domain.ErrInvalidStatusTransition
	}

	return s.orderRepo.UpdateOrderStatus(ctx, id, newStatus, comment)
}

func (s *orderService) ListRestaurantOrders(ctx context.Context, restaurantID uuid.UUID, status *domain.OrderStatus) ([]domain.Order, error) {
	return s.orderRepo.ListOrdersByRestaurant(ctx, restaurantID, status)
}
