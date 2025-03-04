-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS regions (
    id SERIAL PRIMARY KEY,
    RegionId INTEGER,
    RegionName VARCHAR(64),
    UNIQUE (RegionId),
    UNIQUE (RegionName)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS regions;
-- +goose StatementEnd
