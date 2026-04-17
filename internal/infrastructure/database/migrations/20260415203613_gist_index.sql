-- +goose Up
CREATE INDEX idx_subscriptions_period
ON subscriptions
USING GIST (daterange(start_date, COALESCE(end_date, 'infinity')));

-- +goose Down
DROP INDEX IF EXISTS idx_subscriptions_period;
