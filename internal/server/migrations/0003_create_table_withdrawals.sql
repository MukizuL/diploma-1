-- +goose Up
CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL ,
    order_id BIGINT NOT NUll UNIQUE,
    amount NUMERIC(19, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_orders FOREIGN KEY (order_id) REFERENCES orders(order_id)
);

-- +goose Down
DROP TABLE IF EXISTS withdrawals;