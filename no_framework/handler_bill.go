package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/carloscfgos1980/cashier-app/internal/database"
)

type Bills struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Denomination int    `json:"denomination"`
	Quantity     int    `json:"quantity"`
}

func (cfg *apiConfig) handlerBillsCreate(w http.ResponseWriter, r *http.Request) {
	type bill struct {
		Denomination float32 `json:"denomination"`
		Quantity     int32   `json:"quantity"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read request body", err)
		return
	}

	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		respondWithError(w, http.StatusBadRequest, "Request body is empty", nil)
		return
	}

	var bills []bill
	switch body[0] {
	case '[':
		if err := json.Unmarshal(body, &bills); err != nil {
			respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
			return
		}
	case '{':
		var wrapped struct {
			Bills []bill `json:"bills"`
			Bill  []bill `json:"bill"`
		}
		if err := json.Unmarshal(body, &wrapped); err != nil {
			respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
			return
		}
		if len(wrapped.Bills) > 0 {
			bills = wrapped.Bills
		} else {
			bills = wrapped.Bill
		}
	default:
		respondWithError(w, http.StatusBadRequest, "Request body must be an array of bills or an object with bill(s)", nil)
		return
	}

	if len(bills) == 0 {
		respondWithError(w, http.StatusBadRequest, "Request body must include at least one bill", nil)
		return
	}

	for _, b := range bills {
		demon := b.Denomination
		if !validateDenomination(demon) {
			respondWithError(w, http.StatusBadRequest, "Invalid denomination", nil)
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
			dbBill.Quantity += b.Quantity
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
