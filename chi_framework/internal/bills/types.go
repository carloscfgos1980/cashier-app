package bills

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// Define a struct to format the response
type BillResponse struct {
	ID           pgtype.UUID      `json:"id"`
	CreatedAt    pgtype.Timestamp `json:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at"`
	Denomination float32          `json:"denomination"`
	Quantity     int32            `json:"quantity"`
}

// BillRequest represents a bill denomination and its quantity.
type BillRequest struct {
	Denomination float32 `json:"denomination"`
	Quantity     int32   `json:"quantity"`
}

// Transaction represents the request body for the change calculation.
type TransactionRequest struct {
	AmountDue  float32 `json:"amount_due"`
	AmountPaid float32 `json:"amount_paid"`
}

// ChangeLine represents a line in the change response.
type ChangeLine struct {
	Text string `json:"text"`
}
