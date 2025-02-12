-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS regions (
    RegionId INTEGER PRIMARY KEY,
    RegionName VARCHAR(64));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS regions;
-- +goose StatementEnd
