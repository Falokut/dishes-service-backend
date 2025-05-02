-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    username TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    admin BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO
    users (id, username, name)
VALUES (
        '00000000-0000-0000-0000-000000000000',
        'DELETED',
        'DELETED'
    );

CREATE TABLE users_telegrams (
    id uuid REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE UNIQUE,
    chat_id BIGINT NOT NULL UNIQUE,
    telegram_id BIGINT NOT NULL UNIQUE
);

CREATE TABLE restaurants (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE dish (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    -- Цена в минимальных единицах валюты, например, в копейках
    price INT NOT NULL CHECK (price > 0),
    image_id TEXT,
    restaurant_id INT REFERENCES restaurants (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE dish_categories (
    dish_id INT REFERENCES dish (id) ON DELETE CASCADE ON UPDATE CASCADE,
    category_id INT REFERENCES categories (id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (dish_id, category_id)
);


CREATE TABLE orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    payment_method TEXT NOT NULL,
    user_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000' REFERENCES users (id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    total BIGINT NOT NULL CHECK (total > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    wishes TEXT,
    status TEXT NOT NULL
);

CREATE TABLE order_items (
    order_id uuid NOT NULL REFERENCES orders (id) ON DELETE CASCADE ON UPDATE CASCADE,
    dish_id INT NOT NULL REFERENCES dish (id) ON DELETE CASCADE ON UPDATE CASCADE,
    count INT NOT NULL CHECK (count > 0),
    price INT NOT NULL CHECK (price > 0),
    PRIMARY KEY (order_id, dish_id)
);

-- +goose Down
DROP TABLE order_items;

DROP TABLE orders;

DROP TABLE dish_categories;

DROP TABLE dish;

DROP TABLE categories;

DROP TABLE users_telegrams;

DROP TABLE users;