package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/carloscfgos1980/cashier-app/internal/database"
)

func (cfg *apiConfig) handlerBillsCreateUpdate(w http.ResponseWriter, r *http.Request) {
	type bill struct {
		Denomination float32 `json:"denomination"`
		Quantity     int32   `json:"quantity"`
	}

	var bills []bill
	if err := json.NewDecoder(r.Body).Decode(&bills); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	for _, b := range bills {
		demon := b.Denomination
		if !validateDenomination(demon) {
			respondWithError(w, http.StatusBadRequest, "Invalid denomination", nil)
			log.Printf("Invalid denomination: %f", demon)
			return
		}
		// convert the denomination from float32 to int (cents) to avoid floating point precision issues
		demonCents := int32(demon * 100)

		dbBill, err := cfg.db.GetBillByDenomination(r.Context(), demonCents)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get bill by denomination", err)
			return
		}

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
			dbBill.Quantity = b.Quantity
			_, err = cfg.db.UpdateBill(r.Context(), database.UpdateBillParams{
				ID:       dbBill.ID,
				Quantity: dbBill.Quantity,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't update bill", err)
				return
			}
		}
	}

	respondWithJSON(w, http.StatusOK, bills)
}
