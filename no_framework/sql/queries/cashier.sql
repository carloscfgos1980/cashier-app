-- name: CreateBill :one
INSERT INTO bills (id, created_at, updated_at, denomination, quantity)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
) RETURNING *;

-- name: GetBills :many
SELECT * FROM bills ORDER BY denomination DESC;

-- name: GetBillByDenomination :one
SELECT * FROM bills WHERE denomination = $1;  

-- name: UpdateBill :one
UPDATE bills
SET quantity = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


