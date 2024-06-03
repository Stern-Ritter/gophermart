-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL,
    login VARCHAR(30) NOT NULL,
    password VARCHAR(256) NOT NULL,
    CONSTRAINT pk_users PRIMARY KEY(id),
    CONSTRAINT users_login_unique UNIQUE(login)
);

CREATE TYPE  accrual_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS loyalty_points_accrual (
    user_id BIGINT NOT NULL,
    order_number BIGINT NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE,
    status ACCRUAL_STATUS NOT NULL,
    amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    processing_lock BOOL DEFAULT FALSE,
    CONSTRAINT pk_loyalty_points_accrual PRIMARY KEY(user_id, order_number),
    CONSTRAINT loyalty_points_accrual_to_users_fk
    FOREIGN KEY(user_id) REFERENCES users(id),
    CONSTRAINT loyalty_points_accrual_order_number_unique UNIQUE(order_number)
);

CREATE TABLE IF NOT EXISTS loyalty_points_withdrawn (
    user_id BIGINT NOT NULL,
    order_number BIGINT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    CONSTRAINT pk_loyalty_points_withdrawn PRIMARY KEY(user_id, order_number),
    CONSTRAINT loyalty_points_withdrawn_to_users_fk
    FOREIGN KEY(user_id) REFERENCES users(id),
    CONSTRAINT loyalty_points_withdrawn_order_number_unique UNIQUE(order_number)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE loyalty_points_withdrawn;
DROP TABLE loyalty_points_accrual;
DROP TYPE accrual_status;
DROP TABLE users;
-- +goose StatementEnd