-- +goose Up
-- +goose StatementBegin

-- Создание таблицы пользователей
-- Создание таблицы пользователей
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Создание таблицы данных
CREATE TABLE user_data (
    id UUID DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    data_type VARCHAR(50) NOT NULL, -- 'password', 'text', 'binary', 'card'
    encrypted_data BYTEA NOT NULL,  -- зашифрованные данные
    metadata VARCHAR(50) NOT NULL,        -- метаинформация
    version INTEGER DEFAULT 1,       
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индексы для производительности
CREATE INDEX idx_user_data_data_type ON user_data(data_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем индексы
DROP INDEX IF EXISTS idx_user_data_data_type;

-- Удаляем таблицы (в порядке обратной зависимости)
DROP TABLE IF EXISTS user_data;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
