package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/carloscfgos1980/cashier-app/internal/database"
)

// handlerBillsCreateUpdate handles the creation or update of bills in the database.
func (cfg *apiConfig) handlerBillsCreateUpdate(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a slice of Bill structs
	var bills []Bill
	if err := json.NewDecoder(r.Body).Decode(&bills); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// Iterate over the bills and create or update them in the database
	for _, b := range bills {
		demon := b.Denomination
		// Validate the denomination
		if !validateDenomination(demon) {
			respondWithError(w, http.StatusBadRequest, "Invalid denomination", nil)
			log.Printf("Invalid denomination: %f", demon)
			return
		}
		// convert the denomination from float32 to int (cents) to avoid floating point precision issues
		demonCents := int32(demon * 100)
		// Check if the bill already exists in the database
		dbBill, err := cfg.db.GetBillByDenomination(r.Context(), demonCents)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get bill by denomination", err)
			return
		}
		// If the bill does not exist, create it; otherwise, update the quantity
		if errors.Is(err, sql.ErrNoRows) {
			_, err := cfg.db.CreateBill(r.Context(), database.CreateBillParams{
				Denomination: demonCents,
				Quantity:     b.Quantity,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't create bill", err)
				return
			}
			continue
		}

		// check if the bill already exists
		if dbBill.ID.String() != "" {
			// if it exists, update the quantity
			_, err = cfg.db.UpdateBill(r.Context(), database.UpdateBillParams{
				ID:       dbBill.ID,
				Quantity: b.Quantity,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't update bill", err)
				return
			}
		}
	}
	// response with the updated bills
	respondWithJSON(w, http.StatusOK, bills)
}
