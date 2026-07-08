-- +goose Up
ALTER TABLE bills
ADD CONSTRAINT bills_quantity_non_negative CHECK (quantity >= 0) NOT VALID;

-- +goose Down
ALTER TABLE bills
DROP CONSTRAINT IF EXISTS bills_quantity_non_negative;
