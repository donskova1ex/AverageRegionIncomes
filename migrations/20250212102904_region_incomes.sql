-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS region_incomes (
    RegionId INTEGER NOT NULL,
    Year INTEGER NOT NULL,
    Quarter INTEGER NOT NULL,
    Value DECIMAL(18,2) NOT NULL,
    CONSTRAINT PK_RegionIncomes PRIMARY KEY (RegionId, Year, Quarter, Value),
    CONSTRAINT CHK_ValueNonNegative CHECK (Value >= 0)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS region_incomes;
-- +goose StatementEnd
//TODO: Формат данных не отрицательный