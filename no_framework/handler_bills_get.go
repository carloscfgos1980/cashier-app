package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerBillsGet(w http.ResponseWriter, r *http.Request) {
	bills, err := cfg.db.GetBills(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get bills", err)
		return
	}

	type bill struct {
		ID           string  `json:"id"`
		CreatedAt    string  `json:"created_at"`
		UpdatedAt    string  `json:"updated_at"`
		Denomination float32 `json:"denomination"`
		Quantity     int32   `json:"quantity"`
	}

	var response []bill
	for _, b := range bills {
		response = append(response, bill{
			ID:           b.ID.String(),
			CreatedAt:    b.CreatedAt.String(),
			UpdatedAt:    b.UpdatedAt.String(),
			Denomination: float32(b.Denomination) / 100,
			Quantity:     b.Quantity,
		})
	}

	respondWithJSON(w, http.StatusOK, response)
}
