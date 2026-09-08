#!/usr/bin/env bash
set -e

# Цвета для вывода в терминал
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

CORE_URL="${CORE_URL:-http://localhost:8080}"
PARENT_URL="${PARENT_URL:-http://localhost:8081}"
RESTAURANT_ID="a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
USER_ID="d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99"

TOTAL_TESTS=0
PASSED_TESTS=0

pass() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "  [${GREEN}PASS${NC}] $1"
}

fail() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "  [${RED}FAIL${NC}] $1"
    echo -e "         Причина: $2"
    exit 1
}

info() {
    echo -e "\n${BLUE}==>${NC} ${YELLOW}$1${NC}"
}

echo -e "${GREEN}======================================================${NC}"
echo -e "${GREEN}    Запуск E2E тестов платформы «Авито.Кухня»        ${NC}"
echo -e "${GREEN}======================================================${NC}"

# -------------------------------------------------------------
# Тест 1: Healthcheck сервисов
# -------------------------------------------------------------
info "Тест 1: Проверка работоспособности сервисов (Healthchecks)"

CORE_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "${CORE_URL}/health")
if [ "$CORE_HEALTH" -eq 200 ]; then
    pass "Core Service (:8080/health) доступен (200 OK)"
else
    fail "Core Service недоступен" "HTTP код: $CORE_HEALTH"
fi

PARENT_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "${PARENT_URL}/health")
if [ "$PARENT_HEALTH" -eq 200 ]; then
    pass "Parent / Restaurant Service (:8081/health) доступен (200 OK)"
else
    fail "Parent Service недоступен" "HTTP код: $PARENT_HEALTH"
fi

# -------------------------------------------------------------
# Тест 2: Клиентская витрина и поиск заведений
# -------------------------------------------------------------
info "Тест 2: Витрина ресторанов и поиск (GET /api/v1/restaurants)"

RESTS_JSON=$(curl -s "${CORE_URL}/api/v1/restaurants")
if echo "$RESTS_JSON" | grep -q "Авито Бургер"; then
    pass "Список активных ресторанов получен корректно"
else
    fail "Ресторан 'Авито Бургер' не найден в списке" "$RESTS_JSON"
fi

SEARCH_JSON=$(curl -s "${CORE_URL}/api/v1/restaurants?search=Бургер")
if echo "$SEARCH_JSON" | grep -q "Авито Бургер"; then
    pass "Полнотекстовый поиск по названию работает"
else
    fail "Поиск '?search=Бургер' не вернул результат" "$SEARCH_JSON"
fi

# -------------------------------------------------------------
# Тест 3: Меню заведения по категориям
# -------------------------------------------------------------
info "Тест 3: Меню ресторана по категориям (GET /api/v1/restaurants/{id}/menu)"

MENU_JSON=$(curl -s "${CORE_URL}/api/v1/restaurants/${RESTAURANT_ID}/menu")
if echo "$MENU_JSON" | grep -q "Классический Чизбургер"; then
    pass "Меню ресторана успешно получено и содержит категории и блюда"
else
    fail "Меню не содержит 'Классический Чизбургер'" "$MENU_JSON"
fi

# Ищем ID чизбургера и блюда из стоп-листа
ITEM_BURGER_ID=$(python3 -c "
import sys, json
data = json.loads('''$MENU_JSON''')
for cat in data:
    for item in cat.get('items', []):
        if 'Классический Чизбургер' in item['name']:
            print(item['id'])
            sys.exit(0)
")

# -------------------------------------------------------------
# Тест 4: Защита от овербукинга и стоп-листов (409 Conflict)
# -------------------------------------------------------------
info "Тест 4: Проверка бизнес-ограничений (стоп-лист и остатки)"

# Пытаемся заказать 99999 порций (превышение остатка)
OVERBOOK_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${CORE_URL}/api/v1/orders" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: ${USER_ID}" \
  -d '{
    "restaurant_id": "'"${RESTAURANT_ID}"'",
    "customer_name": "Тестер",
    "customer_phone": "+79990000000",
    "delivery_address": "Тестовая ул., 1",
    "items": [{"menu_item_id": "'"${ITEM_BURGER_ID}"'", "quantity": 99999}]
  }')

if [ "$OVERBOOK_STATUS" -eq 409 ]; then
    pass "Превышение остатка корректно отклонено (409 Conflict)"
else
    fail "Ожидался 409 Conflict при превышении остатка" "Получен HTTP: $OVERBOOK_STATUS"
fi

# -------------------------------------------------------------
# Тест 5: Успешное оформление заказа и фиксация цен
# -------------------------------------------------------------
info "Тест 5: Чекаут заказа клиентом (POST /api/v1/orders)"

ORDER_RESP=$(curl -s -X POST "${CORE_URL}/api/v1/orders" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: ${USER_ID}" \
  -d '{
    "restaurant_id": "'"${RESTAURANT_ID}"'",
    "customer_name": "Алексей Иванов",
    "customer_phone": "+79995554433",
    "delivery_address": "Ленинградский пр., д. 39",
    "items": [{"menu_item_id": "'"${ITEM_BURGER_ID}"'", "quantity": 1}]
  }')

ORDER_ID=$(echo "$ORDER_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))")
ORDER_STATUS=$(echo "$ORDER_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('status', ''))")
ORDER_PRICE=$(echo "$ORDER_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('total_price', 0))")

if [ -n "$ORDER_ID" ] && [ "$ORDER_STATUS" == "CREATED" ]; then
    pass "Заказ успешно создан со статусом CREATED (ID: $ORDER_ID)"
else
    fail "Ошибка создания заказа" "$ORDER_RESP"
fi

if [ "$ORDER_PRICE" == "350" ] || [ "$ORDER_PRICE" == "350.0" ]; then
    pass "Снимок цен и общая стоимость рассчитаны точно (350 руб.)"
else
    fail "Неверная итоговая сумма заказа" "$ORDER_PRICE"
fi

# -------------------------------------------------------------
# Тест 6: Асинхронный вебхук в ресторан и продвижение FSM
# -------------------------------------------------------------
info "Тест 6: Асинхронная доставка вебхука и симуляция стадий кухни"

sleep 1
RECEIVED_BY_RESTAURANT=$(curl -s "${PARENT_URL}/api/v1/orders")
if echo "$RECEIVED_BY_RESTAURANT" | grep -q "$ORDER_ID"; then
    pass "Ресторан-эмулятор успешно получил заказ через Webhook"
else
    fail "Заказ не найден в полученных вебхуках ресторана" "$RECEIVED_BY_RESTAURANT"
fi

echo -n "  Ожидание прохождения этапов кухни (ACCEPTED -> COOKING -> READY -> IN_DELIVERY -> DELIVERED)..."
# Симулятор выполняет шаги с паузами (2s + 3s + 3s + 3s + 3s = ~14s)
for i in {1..20}; do
    sleep 1
    echo -n "."
    CURRENT_STATUS=$(curl -s "${CORE_URL}/api/v1/orders/${ORDER_ID}" | python3 -c "import sys, json; print(json.load(sys.stdin).get('status', ''))")
    if [ "$CURRENT_STATUS" == "DELIVERED" ]; then
        break
    fi
done
echo ""

if [ "$CURRENT_STATUS" == "DELIVERED" ]; then
    pass "Заказ успешно завершил полный цикл FSM и перешел в DELIVERED"
else
    fail "Заказ не достиг финального статуса DELIVERED за отведенное время" "Текущий статус: $CURRENT_STATUS"
fi

# -------------------------------------------------------------
# Тест 7: Запрет отмены заказа после начала готовки / доставки
# -------------------------------------------------------------
info "Тест 7: Блокировка отмены заказа после готовки (409 Conflict)"

CANCEL_LATE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${CORE_URL}/api/v1/orders/${ORDER_ID}/cancel" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Передумал"}')

if [ "$CANCEL_LATE_STATUS" -eq 409 ]; then
    pass "Попытка отменить доставленный/готовящийся заказ заблокирована (409 Conflict)"
else
    fail "Ожидался 409 Conflict при отмене DELIVERED заказа" "Получен HTTP: $CANCEL_LATE_STATUS"
fi

# -------------------------------------------------------------
# Тест 8: Отмена до начала готовки и автоматический возврат остатка
# -------------------------------------------------------------
info "Тест 8: Отмена заказа клиентом в статусе CREATED и возврат остатка"

# 1. Запоминаем текущий остаток блюда
STOCK_BEFORE=$(curl -s "${CORE_URL}/api/v1/restaurants/${RESTAURANT_ID}/menu" | python3 -c "
import sys, json
data = json.loads(sys.stdin.read())
for cat in data:
    for item in cat.get('items', []):
        if item['id'] == '$ITEM_BURGER_ID':
            print(item['stock_quantity'])
            sys.exit(0)
")

# 2. Оформляем новый заказ
ORDER_TO_CANCEL=$(curl -s -X POST "${CORE_URL}/api/v1/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id": "'"${RESTAURANT_ID}"'",
    "customer_name": "Тест Отмены",
    "customer_phone": "+79997771122",
    "delivery_address": "Тестовый тупик, 2",
    "items": [{"menu_item_id": "'"${ITEM_BURGER_ID}"'", "quantity": 1}]
  }')

CANCEL_ORDER_ID=$(echo "$ORDER_TO_CANCEL" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))")

# 3. Немедленно отменяем до перехода в COOKING
CANCEL_RESP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${CORE_URL}/api/v1/orders/${CANCEL_ORDER_ID}/cancel" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Ошибся адресом доставки"}')

if [ "$CANCEL_RESP_CODE" -eq 200 ]; then
    pass "Заказ в статусе CREATED успешно отменен клиентом (200 OK)"
else
    fail "Не удалось отменить заказ" "HTTP код: $CANCEL_RESP_CODE"
fi

# 4. Проверяем, что остаток вернулся на склад
STOCK_AFTER=$(curl -s "${CORE_URL}/api/v1/restaurants/${RESTAURANT_ID}/menu" | python3 -c "
import sys, json
data = json.loads(sys.stdin.read())
for cat in data:
    for item in cat.get('items', []):
        if item['id'] == '$ITEM_BURGER_ID':
            print(item['stock_quantity'])
            sys.exit(0)
")

if [ "$STOCK_AFTER" -eq "$STOCK_BEFORE" ]; then
    pass "Списанный остаток товара автоматически возвращен в БД ($STOCK_BEFORE шт.)"
else
    fail "Остаток не вернулся после отмены!" "Было: $STOCK_BEFORE, стало: $STOCK_AFTER"
fi

# -------------------------------------------------------------
# Тест 9: Партнерский контур управления меню и стоп-листами
# -------------------------------------------------------------
info "Тест 9: Партнерский контур: стоп-листы и склад"

# Ставим чизбургер в стоп-лист
STOP_RESP=$(curl -s -o /dev/null -w "%{http_code}" -X PATCH \
  "${CORE_URL}/api/v1/partner/restaurants/${RESTAURANT_ID}/items/${ITEM_BURGER_ID}/availability" \
  -H "Content-Type: application/json" \
  -H "X-Partner-ID: ${RESTAURANT_ID}" \
  -d '{"is_available": false}')

if [ "$STOP_RESP" -eq 200 ]; then
    pass "Партнер перевел блюдо в стоп-лист (PATCH /availability)"
else
    fail "Ошибка изменения доступности" "HTTP код: $STOP_RESP"
fi

# Попытка заказать его клиентом должна завершиться 409 Conflict
TRY_ORDER_STOPPED=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${CORE_URL}/api/v1/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id": "'"${RESTAURANT_ID}"'",
    "customer_name": "Клиент",
    "customer_phone": "+79991112233",
    "delivery_address": "Тверская, 1",
    "items": [{"menu_item_id": "'"${ITEM_BURGER_ID}"'", "quantity": 1}]
  }')

if [ "$TRY_ORDER_STOPPED" -eq 409 ]; then
    pass "Покупка блюда из стоп-листа заблокирована (409 Conflict)"
else
    fail "Ожидался 409 Conflict при заказе блюда из стоп-листа" "Получен: $TRY_ORDER_STOPPED"
fi

# Возвращаем блюдо в меню
curl -s -X PATCH \
  "${CORE_URL}/api/v1/partner/restaurants/${RESTAURANT_ID}/items/${ITEM_BURGER_ID}/availability" \
  -H "Content-Type: application/json" \
  -H "X-Partner-ID: ${RESTAURANT_ID}" \
  -d '{"is_available": true}' > /dev/null
pass "Блюдо возвращено в доступное состояние на витрине"

# -------------------------------------------------------------
# Итог
# -------------------------------------------------------------
echo -e "\n${GREEN}======================================================${NC}"
echo -e "${GREEN}    Все ${PASSED_TESTS}/${TOTAL_TESTS} E2E тестов успешно пройдены!              ${NC}"
echo -e "${GREEN}======================================================${NC}\n"
