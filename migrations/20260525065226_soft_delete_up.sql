-- +goose Up
-- +goose StatementBegin

ALTER TABLE user_data
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL,
ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_user_data_deleted ON user_data(is_deleted);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем индекс
DROP INDEX IF EXISTS idx_user_data_deleted;

-- Удаляем столбцы из таблицы user_data
ALTER TABLE user_data
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS is_deleted;

-- +goose StatementEnd
