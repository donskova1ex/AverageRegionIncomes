-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS region_incomes (
    RegionId INTEGER PRIMARY KEY,
    Year INTEGER,
    Quarter INTEGER,
    Value FLOAT);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS region_incomes;
-- +goose StatementEnd
