-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_subs (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT valid_date_range CHECK (end_date IS NULL OR end_date >= start_date)
    CONSTRAINT fk_user_subs FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
);

CREATE INDEX idx_subs_user_id ON user_subs(user_id);
CREATE INDEX idx_subs_service_name ON user_subs(service_name);
CREATE INDEX idx_subs_dates ON user_subs(start_date, end_date);
CREATE INDEX idx_subs_user_active ON user_subs(user_id) WHERE end_date IS NULL OR end_date >= CURRENT_DATE;

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_subs_updated_at
    BEFORE UPDATE ON user_subs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_user_subs_updated_at ON user_subs;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_subs_user_id;
DROP INDEX IF EXISTS idx_subs_service_name;
DROP INDEX IF EXISTS idx_subs_dates;
DROP INDEX IF EXISTS idx_subs_user_active;

DROP TABLE IF EXISTS user_subs;
-- +goose StatementEnd