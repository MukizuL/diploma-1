-- +goose Up
-- +goose StatementBegin
CREATE TYPE status_enum AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL ,
    order_id BIGINT NOT NUll UNIQUE,
    status status_enum DEFAULT 'NEW' NOT NULL,
    accrual NUMERIC(19, 2) DEFAULT 0 NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX orders_order_id_hash_idx ON orders USING HASH(order_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS status_enum;
-- +goose StatementEnd