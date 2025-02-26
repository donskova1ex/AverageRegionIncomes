-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS regions (
    id SERIAL PRIMARY KEY,
    RegionId INTEGER unique,
    RegionName VARCHAR(64));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS regions;
-- +goose StatementEnd
