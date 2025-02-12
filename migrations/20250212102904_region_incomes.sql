-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS region_incomes (
    RegionId INTEGER PRIMARY KEY,
    Year INTEGER,
    Quarter INTEGER,
    Value DECIMAL(18,2));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS region_incomes;
-- +goose StatementEnd
//TODO: Формат данных не отрицательный