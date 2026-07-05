package bills

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// Define a struct to format the response
type Bill struct {
	ID           pgtype.UUID      `json:"id"`
	CreatedAt    pgtype.Timestamp `json:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at"`
	Denomination float32          `json:"denomination"`
	Quantity     int32            `json:"quantity"`
}
