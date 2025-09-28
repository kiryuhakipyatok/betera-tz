-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tasks(
    id UUID PRIMARY KEY,
    title VARCHAR(150) NOT NULL UNIQUE,
    description VARCHAR(650) NOT NULL,
    status VARCHAR(10) CHECK (status IN ('created', 'done', 'processing')) DEFAULT 'created'
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks
-- +goose StatementEnd
