-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS region_incomes (
    id SERIAL PRIMARY KEY,
    region_id INTEGER NOT NULL,
    year INTEGER NOT NULL,
    quarter INTEGER NOT NULL,
    value DECIMAL(18,2) NOT NULL,
    CONSTRAINT CHK_ValueNonNegative CHECK (value >= 0),
    CONSTRAINT UQ_RegionIncomes UNIQUE (region_id, year, quarter, value)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS region_incomes;
-- +goose StatementEnd