package main

import (
	"net/http"
)

// handlerBillsGet handles the GET /api/bills endpoint. It retrieves all bills from the database and returns them as a JSON response.
func (cfg *apiConfig) handlerBillsGet(w http.ResponseWriter, r *http.Request) {
	// Retrieve all bills from the database
	bills, err := cfg.db.GetBills(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get bills", err)
		return
	}
	// Define a struct to format the response
	type bill struct {
		ID           string  `json:"id"`
		CreatedAt    string  `json:"created_at"`
		UpdatedAt    string  `json:"updated_at"`
		Denomination float32 `json:"denomination"`
		Quantity     int32   `json:"quantity"`
	}
	// Create a slice to hold the formatted bills for the response
	var response []bill
	for _, b := range bills {
		response = append(response, bill{
			ID:           b.ID.String(),
			CreatedAt:    b.CreatedAt.String(),
			UpdatedAt:    b.UpdatedAt.String(),
			Denomination: float32(b.Denomination) / 100, // Convert cents to euros
			Quantity:     b.Quantity,
		})
	}

	// Respond with the formatted bills as JSON
	respondWithJSON(w, http.StatusOK, response)
}
