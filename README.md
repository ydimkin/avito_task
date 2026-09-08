# Авито.Кухня (Avito.Kitchen) — Backend MVP

Бэкенд-сервис платформы заказа и доставки еды из заведений-партнеров для Авито.  
Реализован в рамках тестового задания для стажёра Backend.

---

## 1. Архитектура решения

Система состоит из двух сервисов, запускаемых через Docker Compose:

- **`core-service`** (`:8080`) — ядро платформы: витрина заведений, меню, оформление заказов с транзакционным списанием остатков (`SELECT FOR UPDATE`), контроль жизненного цикла заказа (FSM) и асинхронная отправка вебхуков заведениям.
- **`parent-service`** (`:8081`) — сервис заведения, который принимает входящие вебхуки о заказах, эмулирует этапы приготовления на кухне через фоновые горутины и отправляет обновления статусов обратно в `core-service`.

### Стек:

- **Язык**: Go 1.24 (Gin Web Framework, `log/slog`)
- **База данных**: PostgreSQL 16 (GORM, автоматические миграции)
- **Контракт**: OpenAPI 3.0.3 + встроенный Swagger UI (`/swagger`)
- **Линтер**: `golangci-lint` (конфигурация `.golangci.yml`)
- **Оркестрация**: Docker Compose

---

## 2. Архитектура системы (C4 Model)

Модель архитектуры представлена на уровнях **Level 2 (Containers)** и **Level 3 (Components)**. Все диаграммы выполнены в формате Mermaid.

### 2.1. C4 Level 2: Контейнеры (Containers)

> Исходный файл: [`docs/c4-l2-l3/c4_level2_containers.mmd`](docs/c4-l2-l3/c4_level2_containers.mmd)

```mermaid
flowchart LR
    classDef actor fill:#1A365D,stroke:#0F172A,color:#FFFFFF,stroke-width:2px;
    classDef core fill:#2B6CB0,stroke:#1E4E8C,color:#FFFFFF,stroke-width:2px;
    classDef db fill:#2C5282,stroke:#1A365D,color:#FFFFFF,stroke-width:2px;
    classDef partner fill:#4A5568,stroke:#2D3748,color:#FFFFFF,stroke-width:2px;

    Client["Клиент<br/>(Web / Mobile SPA)"]:::actor
    PartnerStaff["Персонал заведения<br/>(Касса / POS)"]:::actor

    subgraph Platform["Платформа Авито.Кухня"]
        CoreService["Core Service (:8080)<br/>Go 1.24, Gin, GORM<br/>REST API, чекаут, FSM, вебхуки"]:::core
        Database[("PostgreSQL 16 (:5432)<br/>Заведения, меню, заказы,<br/>снимки цен, аудит статусов")]:::db
    end

    subgraph PartnerEnv["Контур заведения-партнера"]
        PartnerService["Restaurant Service (:8081)<br/>parent-service (Go, Gin)<br/>Эмулятор POS и кухни"]:::partner
    end

    Client -->|"Поиск, меню, чекаут, отмена<br/>[HTTP / JSON]"| CoreService
    PartnerStaff -->|"Каталог, стоп-листы, остатки<br/>[HTTP / JSON]"| CoreService

    CoreService -->|"ACID транзакции<br/>SELECT FOR UPDATE"| Database

    CoreService -->|"Новый заказ<br/>[Webhook HTTP POST]"| PartnerService
    PartnerService -->|"Смена статусов FSM<br/>[HTTP PATCH]"| CoreService
```

### 2.2. C4 Level 3: Компоненты Core Service

> Исходный файл: [`docs/c4-l2-l3/c4_level3_components.mmd`](docs/c4-l2-l3/c4_level3_components.mmd)

```mermaid
flowchart TD
    classDef handler fill:#2B6CB0,stroke:#1E4E8C,color:#FFFFFF,stroke-width:2px;
    classDef service fill:#2C5282,stroke:#1A365D,color:#FFFFFF,stroke-width:2px;
    classDef domain fill:#C05621,stroke:#9C4221,color:#FFFFFF,stroke-width:2px;
    classDef repo fill:#2F855A,stroke:#22543D,color:#FFFFFF,stroke-width:2px;
    classDef db fill:#1A365D,stroke:#0F172A,color:#FFFFFF,stroke-width:2px;
    classDef ext fill:#4A5568,stroke:#2D3748,color:#FFFFFF,stroke-width:2px;

    ClientReq["HTTP-запросы клиентов и партнеров"]

    subgraph CoreService["Core Service (:8080)"]

        subgraph Transport["1. Транспортный слой (Gin)"]
            ClientHandler["ClientHandler<br/>GET /restaurants, GET /menu<br/>POST /orders, POST /orders/cancel"]:::handler
            PartnerHandler["PartnerHandler<br/>POST /partner/menu<br/>PATCH /availability, /stock, /status"]:::handler
        end

        subgraph Domain["2. Бизнес-логика (Services & FSM)"]
            OrderService["OrderService<br/>Чекаут, снимки цен, валидация отмен"]:::service
            MenuService["MenuService<br/>Каталог блюд, остатки, стоп-листы"]:::service
            RestService["RestaurantService<br/>Проверка активности заведений"]:::service
            OrderFSM["Order FSM<br/>Матрица переходов статусов"]:::domain
            WebhookService["WebhookService<br/>Асинхронный HTTP-пуш в горутине"]:::service
        end

        subgraph DataAccess["3. Слой доступа к данным (GORM Repositories)"]
            OrderRepo["OrderRepository<br/>SELECT FOR UPDATE, списание/возврат остатков,<br/>логирование order_status_history"]:::repo
            MenuRepo["MenuRepository<br/>Выборка категорий и блюд по ID"]:::repo
            RestRepo["RestaurantRepository<br/>Поиск ресторанов (ILIKE)"]:::repo
        end
    end

    Database[("PostgreSQL 16 (:5432)<br/>Таблицы каталога, заказов и аудита")]:::db
    PartnerService["Restaurant Service (:8081)<br/>POST /webhook/orders"]:::ext

    ClientReq --> ClientHandler & PartnerHandler

    ClientHandler --> OrderService & MenuService & RestService
    PartnerHandler --> OrderService & MenuService

    OrderService --> OrderFSM
    OrderService --> OrderRepo & MenuRepo
    OrderService -.->|"Фоновый вызов"| WebhookService

    MenuService --> MenuRepo
    RestService --> RestRepo

    WebhookService -->|"HTTP POST"| PartnerService

    OrderRepo -->|"ACID транзакции (SELECT FOR UPDATE)"| Database
    MenuRepo --> Database
    RestRepo --> Database
```

---

## 3. Схема базы данных (ERD)

> Исходный файл: [`docs/erd/database_schema.mmd`](docs/erd/database_schema.mmd)

```mermaid
erDiagram
    RESTAURANTS ||--o{ MENU_CATEGORIES : "has"
    RESTAURANTS ||--o{ MENU_ITEMS : "has"
    RESTAURANTS ||--o{ ORDERS : "receives"
    MENU_CATEGORIES ||--o{ MENU_ITEMS : "groups"
    ORDERS ||--o{ ORDER_ITEMS : "contains"
    ORDERS ||--o{ ORDER_STATUS_HISTORY : "tracks"
    MENU_ITEMS ||--o{ ORDER_ITEMS : "references"

    RESTAURANTS {
        uuid id PK
        string name
        text description
        string address
        string webhook_url
        boolean is_active
        timestamp created_at
    }

    MENU_CATEGORIES {
        uuid id PK
        uuid restaurant_id FK
        string name
        int sort_order
    }

    MENU_ITEMS {
        uuid id PK
        uuid restaurant_id FK
        uuid category_id FK
        string name
        numeric price
        int stock_quantity
        boolean is_available
    }

    ORDERS {
        uuid id PK
        uuid user_id
        uuid restaurant_id FK
        string customer_name
        string customer_phone
        string delivery_address
        string status
        numeric total_price
        timestamp created_at
    }

    ORDER_ITEMS {
        uuid id PK
        uuid order_id FK
        uuid menu_item_id FK
        string name_at_moment
        numeric price_at_moment
        int quantity
        numeric subtotal
    }

    ORDER_STATUS_HISTORY {
        uuid id PK
        uuid order_id FK
        string status
        text comment
        timestamp created_at
    }
```

### Ключевые решения схемы:

1. **Снимки цен (`price_at_moment`)**: в `order_items` фиксируется название и цена на момент совершения заказа. Изменение цен рестораном не влияет на оформленные заказы.
2. **Защита от овербукинга (`SELECT FOR UPDATE`)**: при создании заказа строки блюд блокируются в транзакции. При нехватке — откат с ошибкой `409 Conflict`.
3. **Аудит переходов (`order_status_history`)**: все смены статусов заказа логируются со штампом времени и комментарием.

---

## 4. Сценарии взаимодействия (CJM)

Подробное описание матриц шагов, возможных ошибок и логики FSM вынесено в документ [**`docs/cjm/CJM.md`**](docs/cjm/CJM.md).

### 4.1. CJM Клиента (Customer Journey Map)

> Исходный файл: [`docs/cjm/client_journey.mmd`](docs/cjm/client_journey.mmd)

```mermaid
journey
    title CJM: Путь клиента в «Авито.Кухня»
    section 1. Поиск заведения
      Открытие витрины: 5: Клиент, Web
      Поиск ресторана по названию: 4: Клиент, Core Service
      Заведение временно закрыто: 2: Клиент
      Выбор открытого ресторана: 5: Клиент, Web
    section 2. Выбор блюд
      Просмотр категорий меню: 5: Клиент, Web
      Блюдо в стоп-листе: 2: Клиент, Web
      Выбор доступной альтернативы: 4: Клиент, Web
      Добавление позиций в корзину: 5: Клиент, Web
    section 3. Оформление
      Заполнение адреса и телефона: 4: Клиент, Web
      Отправка заказа на сервер: 5: Клиент, Core Service
      Нехватка остатка (ошибка 409 Conflict): 1: Клиент, Core Service
      Корректировка количества в корзине: 4: Клиент, Web
      Успешное создание заказа (CREATED): 5: Клиент, Core Service
    section 4. Приготовление
      Ресторан подтвердил заказ (ACCEPTED): 5: Клиент, Ресторан
      Отмена до начала готовки (успех): 4: Клиент, Core Service
      Старт готовки шеф-поваром (COOKING): 5: Ресторан
      Попытка отмены во время готовки (отказ 409): 2: Клиент, Core Service
    section 5. Доставка
      Заказ упакован (READY_FOR_DELIVERY): 5: Ресторан
      Курьер забрал заказ (IN_DELIVERY): 4: Клиент, Курьер
      Вручение заказа клиенту (DELIVERED): 5: Клиент, Курьер
```

#### Обработка барьеров клиента:

- **Заведение закрыто**: фильтрация `is_active = true`, поиск по подстроке `ILIKE %search%`.
- **Стоп-лист**: флаг `is_available = false` отключает добавление в корзину на фронтенде и блокирует чекаут на бэкенде.
- **Гонка остатков**: блокировка `SELECT FOR UPDATE` гарантирует, что если два клиента одновременно заказывают последний товар, первый оформит заказ, а второй получит `409 Conflict` без списания средств.
- **Окно отмены**: отмена разрешена до начала готовки (`CREATED`, `ACCEPTED`) с автоматическим возвратом остатков на склад. При переходе в `COOKING` отмена блокируется (`409 Conflict`), защищая ресторан от убытков.

---

### 4.2. CJM Заведения-партнера (Partner Journey Map)

> Исходный файл: [`docs/cjm/partner_journey.mmd`](docs/cjm/partner_journey.mmd)

```mermaid
journey
    title CJM: Путь заведения-партнера
    section 1. Каталог и меню
      Вход в партнерский интерфейс: 5: Партнер, POS
      Выгрузка меню и цен (POST /menu): 5: Партнер, Core Service
      Проверка позиций на витрине: 4: Партнер, Web
    section 2. Оперативное управление
      Закончился ингредиент на кухне: 2: Шеф-повар
      Мгновенный ввод в стоп-лист (PATCH /availability): 5: Шеф-повар, Core Service
      Пополнение остатков (PATCH /stock): 5: Партнер, Core Service
    section 3. Прием заказа
      Получение вебхука (POST /webhook/orders): 5: Core Service, POS
      Подтверждение заказа (ACCEPTED): 5: Шеф-повар, Core Service
    section 4. Приготовление
      Старт приготовления на кухне (COOKING): 5: Шеф-повар, Core Service
      Блокировка отмены клиентом: 5: Ресторан, Core Service
      Упаковка готового блюда (READY_FOR_DELIVERY): 5: Повар, Core Service
    section 5. Передача в доставку
      Передача пакета курьеру (IN_DELIVERY): 5: Шеф-повар, Курьер
      Отметка о доставке клиенту (DELIVERED): 5: Курьер, Core Service
```

#### Особенности интеграции заведения:

- **Управление меню**: пакетная перезапись каталога через `POST .../menu` в транзакции.
- **Экстренный стоп-лист**: `PATCH .../availability` мгновенно скрывает блюдо из выдачи клиентам.
- **Асинхронные вебхуки**: отправка из фоновой горутины не блокирует чекаут клиента.
- **Резервный механизм**: при сетевых сбоях ресторан может запросить очередь активных заказов через поллинг `GET /api/v1/partner/restaurants/{id}/orders?status=CREATED`.

---

## 5. Спецификация API

- **OpenAPI Контракт**: [`api/openapi.yml`](api/openapi.yml) (OpenAPI 3.0.3).
- **Swagger UI**: доступен по адресу `http://localhost:8080/swagger` (или `/docs`).

> **Инженерное решение по кодогенерации**: На начальном этапе рассматривалось использование генерации `openapi.yml` через Swagger-аннотации в коде (`swag`). Однако в процессе это сильно вызывало постоянную путаницу. Поэтому было принято решение отказаться от кодогенератора и описать контракт вручную в [`api/openapi.yml`](api/openapi.yml) (Contract-First подход), подключив его напрямую к встроенному Swagger UI. Это сохранило чистоту исходного кода и дало полный контроль над деталями контракта и статус-кодами.

### Ключевые эндпоинты:

| Метод   | Путь                                                           | Описание                                     | Заголовок      |
| ------- | -------------------------------------------------------------- | -------------------------------------------- | -------------- |
| `GET`   | `/health`                                                      | Проверка работоспособности сервиса           | —              |
| `GET`   | `/api/v1/restaurants`                                          | Список активных заведений (поиск `?search=`) | —              |
| `GET`   | `/api/v1/restaurants/{id}/menu`                                | Меню заведения по категориям                 | —              |
| `POST`  | `/api/v1/orders`                                               | Оформление заказа клиентом                   | `X-User-ID`    |
| `GET`   | `/api/v1/orders/{id}`                                          | Детали заказа и история статусов             | `X-User-ID`    |
| `POST`  | `/api/v1/orders/{id}/cancel`                                   | Отмена заказа клиентом (до `COOKING`)        | `X-User-ID`    |
| `POST`  | `/api/v1/partner/restaurants/{id}/menu`                        | Синхронизация каталога заведения             | `X-Partner-ID` |
| `PATCH` | `/api/v1/partner/restaurants/{id}/items/{itemId}/availability` | Управление стоп-листом                       | `X-Partner-ID` |
| `PATCH` | `/api/v1/partner/restaurants/{id}/items/{itemId}/stock`        | Изменение остатка блюда                      | `X-Partner-ID` |
| `PATCH` | `/api/v1/partner/orders/{id}/status`                           | Смена статуса заказа по FSM                  | `X-Partner-ID` |

---

## 6. Запуск и проверка

### Запуск через Docker Compose

```bash
docker compose up --build -d
```

Порты сервисов:

- **Core Service**: `http://localhost:8080` (Swagger: `http://localhost:8080/swagger`)
- **Parent Service (Эмулятор)**: `http://localhost:8081`
- **PostgreSQL**: `localhost:5432`

### Команды Makefile

Посмотреть все доступные команды:

```bash
make help
```

```bash
Avito Kitchen Service — команды:

  make build           -- сборка сервисов в бинарники в директории ./bin
  make test            -- запуск модульных тестов с детектором гонок
  make e2e             -- запуск сквозного E2E тестирования (bash сценарии)
  make lint            -- запуск линтера
  make up              -- запуск проекта в Docker Compose
  make down            -- остановка проекта в Docker Compose
  make restart         -- перезапуск проекта в Docker Compose
  make logs            -- просмотр всех логов Docker Compose
  make logs-core       -- просмотр логов core-service
  make logs-parent     -- просмотр логов parent-service
  make clean           -- очистка директории ./bin/
```

---

## 7. Использование ИИ и архитектурный план

Подробный мастер-промпт, декомпозиция архитектурных решений и этапы разработки зафиксированы в документе [**`GEMINI.md`**](GEMINI.md).
