-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS region_incomes (
    id SERIAL PRIMARY KEY,
    RegionId INTEGER NOT NULL,
    Year INTEGER NOT NULL,
    Quarter INTEGER NOT NULL,
    Value DECIMAL(18,2) NOT NULL,
    CONSTRAINT CHK_ValueNonNegative CHECK (Value >= 0),
    CONSTRAINT UQ_RegionIncomes UNIQUE (RegionId, Year, Quarter, Value)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS region_incomes;
-- +goose StatementEnd