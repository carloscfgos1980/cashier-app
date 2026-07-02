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

// Change represents the request body for the change calculation.
type Change struct {
	AmountDue  float32 `json:"amount_due"`
	AmountPaid float32 `json:"amount_paid"`
}

// Bill represents a bill denomination and its quantity.
type Bill struct {
	Denomination float32 `json:"denomination"`
	Quantity     int32   `json:"quantity"`
}

// ChangeLine represents a line in the change response.
type ChangeLine struct {
	Text string `json:"text"`
}

// handlerGetChange handles the GET /change endpoint to calculate and return the change.
func (cfg *apiConfig) handlerGetChange(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a Change struct.
	var change Change
	if err := json.NewDecoder(r.Body).Decode(&change); err != nil {
		log.Printf("Error decoding request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}
	// Validate the input amounts.
	if change.AmountPaid < change.AmountDue {
		respondWithError(w, http.StatusBadRequest, "Amount paid is less than amount due", nil)
		return
	}
	// Calculate the change amount.
	changeAmount := change.AmountPaid - change.AmountDue
	// Get the total amount of bills in the store to ensure we have enough change.
	totalInStoreCents, err := cfg.db.GetBillsTotalAmount(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			totalInStoreCents = 0
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get total amount from bills", err)
			return
		}
	}
	// Check if we have enough change in the store.
	if float64(changeAmount*100) > float64(totalInStoreCents) {
		respondWithError(w, http.StatusBadRequest, "Not enough change in store", nil)
		return
	}

	// Convert changeAmount to cents to avoid floating point precision issues.
	changeAmountCents := int32(math.Round(float64(changeAmount * 100)))
	// Get the available bills from the database.
	bills, err := cfg.db.GetBills(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get bills", err)
		return
	}
	// Calculate the change to be given using the available bills.
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
		// Get the current bill from the database by denomination.
		dbBill, err := cfg.db.GetBillByDenomination(r.Context(), int32(bill.Denomination*100))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get bill by denomination", err)
			return
		}
		// Calculate the new quantity of the bill after giving change.
		newQuantity := dbBill.Quantity - bill.Quantity
		if newQuantity < 0 {
			respondWithError(w, http.StatusBadRequest, "Not enough bills in store for denomination", nil)
			return
		}
		// Update the bill quantity in the database.
		_, err = cfg.db.UpdateBill(r.Context(), database.UpdateBillParams{
			ID:       dbBill.ID,
			Quantity: newQuantity,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't update bill quantity", err)
			return
		}
	}
	// Respond with the formatted change response.
	respondWithJSON(w, http.StatusOK, formatChangeResponse(changeBills))

}

// formatChangeResponse formats the change bills into a slice of ChangeLine for the response.
func formatChangeResponse(changeBills []Bill) []ChangeLine {
	lines := make([]ChangeLine, 0, len(changeBills))
	for _, bill := range changeBills {
		billTotal := float64(bill.Denomination) * float64(bill.Quantity)
		lines = append(lines, ChangeLine{Text: formatChangeLine(bill.Quantity, bill.Denomination, billTotal)})
	}
	return lines
}

// formatChangeLine formats a single line of change information.
func formatChangeLine(quantity int32, denomination float32, total float64) string {
	quantityText := strconv.FormatInt(int64(quantity), 10)
	denominationText := formatEuroAmount(float64(denomination))
	totalText := formatEuroAmount(total)
	return quantityText + " x " + denominationText + " = " + totalText
}

// formatEuroAmount formats a float64 value as a Euro currency string.
func formatEuroAmount(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 2, 64)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	return "€" + formatted
}

// calculateChange calculates the change to be given based on the available bills.
func calculateChange(changeAmountCents int32, bills []database.Bill) ([]Bill, error) {
	changeBills := make([]Bill, 0)
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
			changeBills = append(changeBills, Bill{
				Denomination: float32(billValue) / 100,
				Quantity:     numBills,
			})
			changeAmountCents -= numBills * billValue
		}

	}
	return changeBills, nil
}
