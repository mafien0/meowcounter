-- +goose Up
CREATE TABLE counter(
	name TEXT PRIMARY KEY,
	count INTEGER
);

-- +goose Down
DROP TABLE counter;
