-- main tables
CREATE TABLE tariffs
(
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(255) UNIQUE NOT NULL,
    price         BIGINT       NOT NULL CHECK (price >= 0),
    duration_days INTEGER      NOT NULL CHECK (duration_days > 0)
);

-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
CREATE TABLE resources
(
    id          SERIAL PRIMARY KEY,
    chat_id     BIGINT UNIQUE NOT NULL,
    description TEXT
);

-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
CREATE TABLE promocodes
(
    id         SERIAL PRIMARY KEY,
    code       VARCHAR(50) NOT NULL UNIQUE,
    discount   INTEGER     NOT NULL CHECK (discount >= 0 AND discount <= 100),
    expires_at TIMESTAMPTZ,
    used_count INTEGER DEFAULT 0
);

CREATE INDEX idx_promocodes_code ON promocodes (code);

-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
CREATE TABLE requisites
(
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(255) NOT NULL,
    link    TEXT UNIQUE NOT NULL,
    content TEXT NOT NULL,
    photo   BYTEA
);

CREATE TABLE users
(
    id           SERIAL PRIMARY KEY,
    tg_id        BIGINT NOT NULL UNIQUE,
    username     VARCHAR(255),
    first_time   TIMESTAMPTZ DEFAULT NOW(),
    total_sub    INTEGER     DEFAULT 0,
    contains_sub BOOLEAN     DEFAULT FALSE,
    promocode_id INTEGER REFERENCES promocodes (id)
);

CREATE INDEX idx_users_promocode_id ON users (promocode_id);
CREATE INDEX idx_users_tg_id ON users (tg_id);

-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
CREATE TABLE payments
(
    id            SERIAL PRIMARY KEY,
    user_tg_id       INTEGER     NOT NULL REFERENCES users (tg_id) ON DELETE CASCADE,
    amount        BIGINT      NOT NULL CHECK (amount >= 0),
    timestamp     TIMESTAMPTZ DEFAULT NOW(),
    status        VARCHAR(50) NOT NULL,
    receipt_photo BYTEA
);

CREATE INDEX idx_payments_user_id ON payments (user_tg_id);
CREATE INDEX idx_payments_status ON payments (status);

-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
CREATE TABLE subscriptions
(
    id         SERIAL PRIMARY KEY,
    user_tg_id    INTEGER     NOT NULL REFERENCES users (tg_id) ON DELETE CASCADE,
    tariff_id  INTEGER     NOT NULL REFERENCES tariffs (id),
    start_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_date   TIMESTAMPTZ,
    status     VARCHAR(50) NOT NULL
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_tg_id);
CREATE INDEX idx_subscriptions_tariff_id ON subscriptions (tariff_id);
CREATE INDEX idx_subscriptions_status ON subscriptions (status);


-- linking tables
CREATE TABLE tariffs_resources
(
    tariff_id   INTEGER NOT NULL REFERENCES tariffs (id) ON DELETE CASCADE,
    resource_id INTEGER NOT NULL REFERENCES resources (id) ON DELETE CASCADE,
    PRIMARY KEY (tariff_id, resource_id)
);

CREATE INDEX idx_tariffs_resources_tariff_id ON tariffs_resources (tariff_id);
CREATE INDEX idx_tariffs_resources_resource_id ON tariffs_resources (resource_id);

CREATE TABLE promocodes_tariffs
(
    promocode_id INTEGER NOT NULL REFERENCES promocodes (id) ON DELETE CASCADE,
    tariff_id    INTEGER NOT NULL REFERENCES tariffs (id) ON DELETE CASCADE,
    PRIMARY KEY (promocode_id, tariff_id)
);

CREATE INDEX idx_promocodes_tariffs_promocode_id ON promocodes_tariffs (promocode_id);
CREATE INDEX idx_promocodes_tariffs_tariff_id ON promocodes_tariffs (tariff_id);