-- +goose Up
CREATE TABLE bills (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    denomination INT UNIQUE NOT NULL,
    quantity INT NOT NULL
);

-- +goose Down
DROP TABLE bills;