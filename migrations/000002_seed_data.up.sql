-- Добавление тестового ресторана 1
INSERT INTO restaurants (id, name, description, address, webhook_url, is_active)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Авито Бургер',
    'Лучшие сочные бургеры и закуски от шефа',
    'г. Москва, ул. Лесная, д. 7',
    'http://localhost:8081/webhook/orders',
    true
) ON CONFLICT (id) DO NOTHING;

-- Добавление категорий для ресторана 1
INSERT INTO menu_categories (id, restaurant_id, name, sort_order)
VALUES 
    ('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Бургеры', 1),
    ('b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Напитки', 2)
ON CONFLICT (id) DO NOTHING;

-- Добавление блюд для ресторана 1
INSERT INTO menu_items (id, restaurant_id, category_id, name, description, price, stock_quantity, is_available)
VALUES
    ('c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Классический Чизбургер', 'Говяжья котлета, сыр чеддер, маринованные огурцы, фирменный соус', 350.00, 50, true),
    ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Двойной Бекон Бургер', 'Две сочные котлеты, хрустящий бекон, карамелизованный лук', 490.00, 1, true),
    ('c3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Трюфельный Бургер (Стоп-лист)', 'Премиальная мраморная говядина и трюфельный соус', 650.00, 0, false),
    ('c4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Морс Клюквенный 0.5', 'Домашний морс из свежей таежной клюквы', 150.00, 100, true)
ON CONFLICT (id) DO NOTHING;

-- Добавление тестового ресторана 2
INSERT INTO restaurants (id, name, description, address, webhook_url, is_active)
VALUES (
    'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'Суши & Роллы Авито',
    'Свежая рыба и классические японские рецепты',
    'г. Москва, ул. Тверская, д. 12',
    'http://localhost:8081/webhook/orders',
    true
) ON CONFLICT (id) DO NOTHING;

-- Добавление категорий для ресторана 2
INSERT INTO menu_categories (id, restaurant_id, name, sort_order)
VALUES 
    ('b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Сеты', 1)
ON CONFLICT (id) DO NOTHING;

-- Добавление блюд для ресторана 2
INSERT INTO menu_items (id, restaurant_id, category_id, name, description, price, stock_quantity, is_available)
VALUES
    ('c5eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Сет Филадельфия', '32 кусочка свежих роллов с лососем и сливочным сыром', 1290.00, 20, true)
ON CONFLICT (id) DO NOTHING;
