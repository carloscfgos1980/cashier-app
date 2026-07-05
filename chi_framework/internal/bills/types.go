package bills

// Define a struct to format the response
type Bill struct {
	ID           string  `json:"id"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	Denomination float32 `json:"denomination"`
	Quantity     int32   `json:"quantity"`
}
