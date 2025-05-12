-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS region_incomes (
    id SERIAL PRIMARY KEY,
    region_id INTEGER NOT NULL,
    year INTEGER NOT NULL,
    quarter INTEGER NOT NULL,
    value DECIMAL(18,2) NOT NULL,
    loaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT CHK_ValueNonNegative CHECK (value >= 0),
    CONSTRAINT UQ_RegionIncomes UNIQUE (region_id, year, quarter, value)
    );
CREATE INDEX IF NOT EXISTS idx_region_incomes_region_id ON region_incomes (region_id);
CREATE INDEX IF NOT EXISTS idx_region_incomes_region_year ON region_incomes (region_id, year);
CREATE INDEX IF NOT EXISTS idx_region_incomes_region_quarter_year ON region_incomes (region_id, quarter, year);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_region_incomes_region_id;
DROP INDEX IF EXISTS idx_region_incomes_region_year;
DROP INDEX IF EXISTS idx_region_incomes_region_quarter_year;
DROP TABLE IF EXISTS region_incomes;
-- +goose StatementEnd