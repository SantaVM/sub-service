-- +goose Up
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,

    CONSTRAINT check_end_date_after_start_date 
        CHECK (end_date IS NULL OR end_date > start_date),

    CONSTRAINT no_overlapping_subscriptions
    EXCLUDE USING gist (
        user_id WITH =,
        service_name WITH =,
        daterange(start_date, COALESCE(end_date, 'infinity'::date), '[)') WITH &&
    )
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions(start_date);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
