package domain

import "errors"

var (
	// ErrNotFoundConfig возвращается, когда файл конфигурации не найден
	ErrNotFoundConfig = errors.New("configuration file not found")

	// ErrNotFound возвращается, когда запрашиваемая сущность не найдена
	ErrNotFound = errors.New("entity not found")

	// ErrItemInStopList возвращается при попытке заказать блюдо, находящееся в стоп-листе
	ErrItemInStopList = errors.New("menu item is currently unavailable (in stop-list)")

	// ErrCannotCancelOrder возвращается, если заказ нельзя отменить на текущем этапе
	ErrCannotCancelOrder = errors.New("order cannot be cancelled at this stage")

	// ErrInvalidStatusTransition возвращается при попытке недопустимого перехода статуса
	ErrInvalidStatusTransition = errors.New("invalid order status transition")

	// ErrInvalidInput возвращается при невалидных входных параметрах запроса
	ErrInvalidInput = errors.New("invalid input data")

	// ErrRestaurantInactive возвращается при попытке взаимодействия с неактивным рестораном
	ErrRestaurantInactive = errors.New("restaurant is inactive")

	// ErrEmptyOrder возвращается при попытке создать пустой заказ
	ErrEmptyOrder = errors.New("order must contain at least one item")

	// ErrOutOfStock возвращается при попытке заказать товар, количество которого меньше запрашиваемого
	ErrOutOfStock = errors.New("insufficient stock for menu item")

	// ErrPartnerIDMismatch возвращается, если X-Partner-ID не совпадает с id ресторана в пути
	ErrPartnerIDMismatch = errors.New("X-Partner-ID header does not match restaurant id in URL path")

	// ErrRestaurantIDRequired возвращается, если id ресторана не передан
	ErrRestaurantIDRequired = errors.New("restaurant id must be provided in URL path or X-Partner-ID header")

	// ErrWebhookFailed возвращается при ошибке отправки вебхука в ресторан
	ErrWebhookFailed = errors.New("webhook delivery failed with non-2xx status")
)
