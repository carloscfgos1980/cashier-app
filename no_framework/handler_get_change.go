package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/carloscfgos1980/cashier-app/internal/database"
)

type Change struct {
	AmountDue  float32 `json:"amount_due"`
	AmountPaid float32 `json:"amount_paid"`
}

type ChangeBill struct {
	Denomination float32 `json:"denomination"`
	Quantity     int32   `json:"quantity"`
}

type ChangeLine struct {
	Text string `json:"text"`
}

func (cfg *apiConfig) handlerGetChange(w http.ResponseWriter, r *http.Request) {
	var change Change
	if err := json.NewDecoder(r.Body).Decode(&change); err != nil {
		log.Printf("Error decoding request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}

	if change.AmountPaid < change.AmountDue {
		respondWithError(w, http.StatusBadRequest, "Amount paid is less than amount due", nil)
		return
	}

	changeAmount := change.AmountPaid - change.AmountDue

	totalInStoreCents, err := cfg.db.GetBillsTotalAmount(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			totalInStoreCents = 0
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get total amount from bills", err)
			return
		}
	}

	if float64(changeAmount*100) > float64(totalInStoreCents) {
		respondWithError(w, http.StatusBadRequest, "Not enough change in store", nil)
		return
	}

	// Convert changeAmount to cents to avoid floating point precision issues.
	changeAmountCents := int32(math.Round(float64(changeAmount * 100)))

	bills, err := cfg.db.GetBills(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get bills", err)
		return
	}

	changeBills, err := calculateChange(changeAmountCents, bills)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't calculate change", err)
		return
	}
	// If changeBills is empty, it means we couldn't make the exact change with the available bills.
	if len(changeBills) == 0 {
		respondWithError(w, http.StatusBadRequest, "Couldn't make exact change with available bills", nil)
		return
	}
	// update the quantities of the bills in the database
	for _, bill := range changeBills {
		dbBill, err := cfg.db.GetBillByDenomination(r.Context(), int32(bill.Denomination*100))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get bill by denomination", err)
			return
		}
		newQuantity := dbBill.Quantity - bill.Quantity
		if newQuantity < 0 {
			respondWithError(w, http.StatusBadRequest, "Not enough bills in store for denomination", nil)
			return
		}
		_, err = cfg.db.UpdateBill(r.Context(), database.UpdateBillParams{
			ID:       dbBill.ID,
			Quantity: newQuantity,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't update bill quantity", err)
			return
		}
	}

	respondWithJSON(w, http.StatusOK, formatChangeResponse(changeBills))

}

func formatChangeResponse(changeBills []ChangeBill) []ChangeLine {
	lines := make([]ChangeLine, 0, len(changeBills))
	for _, bill := range changeBills {
		billTotal := float64(bill.Denomination) * float64(bill.Quantity)
		lines = append(lines, ChangeLine{Text: formatChangeLine(bill.Quantity, bill.Denomination, billTotal)})
	}
	return lines
}

func formatChangeLine(quantity int32, denomination float32, total float64) string {
	quantityText := strconv.FormatInt(int64(quantity), 10)
	denominationText := formatEuroAmount(float64(denomination))
	totalText := formatEuroAmount(total)
	return quantityText + " x " + denominationText + " = " + totalText
}

func formatEuroAmount(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 2, 64)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	return "€" + formatted
}

func calculateChange(changeAmountCents int32, bills []database.Bill) ([]ChangeBill, error) {
	changeBills := make([]ChangeBill, 0)
	for _, bill := range bills {
		if changeAmountCents <= 0 {
			break
		}
		billValue := bill.Denomination
		billQuantity := bill.Quantity

		if billValue <= changeAmountCents && billQuantity > 0 {
			numBills := changeAmountCents / billValue
			if numBills > billQuantity {
				numBills = billQuantity
			}
			changeBills = append(changeBills, ChangeBill{
				Denomination: float32(billValue) / 100,
				Quantity:     numBills,
			})
			changeAmountCents -= numBills * billValue
		}

	}
	return changeBills, nil
}
