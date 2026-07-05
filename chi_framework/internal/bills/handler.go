package bills

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/carloscfgos1980/cashier-app/internal/database"
	"github.com/carloscfgos1980/cashier-app/internal/json"
)

// handler is the HTTP handler for users endpoints
type handler struct {
	service Service
}

// NewHandler creates a new handler for users endpoints
func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetBills(w http.ResponseWriter, r *http.Request) {
	bills, err := h.service.GetBills(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var response []BillResponse
	for _, bill := range bills {
		response = append(response, BillResponse{
			ID:           bill.ID,
			CreatedAt:    bill.CreatedAt,
			UpdatedAt:    bill.UpdatedAt,
			Denomination: float32(bill.Denomination) / 100, // convert from cents to dollars,
			Quantity:     bill.Quantity,
		})
	}
	if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) BillsCreateUpdate(w http.ResponseWriter, r *http.Request) {
	var bills []BillRequest
	if err := json.ReadJSON(r, &bills); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, b := range bills {
		demon := b.Denomination
		if !ValidateDenomination(demon) {
			http.Error(w, "Invalid denomination", http.StatusBadRequest)
			return
		}
		// convert the denomination from float32 to int (cents) to avoid floating point precision issues
		demonCents := int32(demon * 100)
		dbBill, err := h.service.GetBillByDenomination(r.Context(), demonCents)
		if err != nil {
			if err.Error() == "no rows in result set" {
				// bill does not exist, create it
				_, err := h.service.CreateBill(r.Context(), demonCents, b.Quantity)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// check if the bill already exists
		if dbBill.ID.Valid {
			// bill exists, update it
			_, err := h.service.UpdateBill(r.Context(), dbBill.ID, b.Quantity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err := json.WriteJSON(w, http.StatusOK, bills); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func (h *handler) GetChange(w http.ResponseWriter, r *http.Request) {
	var transaction TransactionRequest
	if err := json.ReadJSON(r, &transaction); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Validate the input amounts.
	if transaction.AmountPaid < transaction.AmountDue {
		http.Error(w, "Amount paid must be greater than or equal to amount due", http.StatusBadRequest)
		return
	}
	// Calculate the change amount.
	changeAmount := transaction.AmountPaid - transaction.AmountDue
	// Convert the change amount to cents to avoid floating point precision issues.
	changeAmountCents := int32(changeAmount * 100)

	// Get the total amount of bills in the database.
	totalAmount, err := h.service.GetBillsTotalAmount(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if there are enough bills to give change.
	if int64(changeAmountCents) > totalAmount {
		http.Error(w, "Not enough bills to give change", http.StatusBadRequest)
		return
	}
	bills, err := h.service.GetBills(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Calculate the change to be given using the available bills.
	changeBills, err := calculateChange(changeAmountCents, bills)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(changeBills) == 0 {
		http.Error(w, "Not enough bills to give change", http.StatusBadRequest)
		return
	}
	// update the quantities of the bills in the database
	for _, bill := range changeBills {
		dbBill, err := h.service.GetBillByDenomination(r.Context(), int32(bill.Denomination*100))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newQuantity := dbBill.Quantity - bill.Quantity
		if newQuantity < 0 {
			http.Error(w, "Not enough bills to give change", http.StatusBadRequest)
			return
		}
		_, err = h.service.UpdateBill(r.Context(), dbBill.ID, newQuantity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// Format the change response.
	changeResponse := formatChangeResponse(changeBills)
	if err := json.WriteJSON(w, http.StatusOK, changeResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func ValidateDenomination(denomination float32) bool {
	switch denomination {
	case 100.00, 50.00, 20.00, 10.00, 5.00, 1.00, 0.50, 0.20, 0.10, 0.05, 0.01:
		return true
	default:
		return false
	}
}

// formatChangeResponse formats the change bills into a slice of ChangeLine for the response.
func formatChangeResponse(changeBills []BillResponse) []ChangeLine {
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
func calculateChange(changeAmountCents int32, bills []database.Bill) ([]BillResponse, error) {
	changeBills := make([]BillResponse, 0)
	sortedBills := append([]database.Bill(nil), bills...)
	sort.Slice(sortedBills, func(i, j int) bool {
		return sortedBills[i].Denomination > sortedBills[j].Denomination
	})

	for _, bill := range sortedBills {
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
			changeBills = append(changeBills, BillResponse{
				Denomination: float32(billValue) / 100,
				Quantity:     numBills,
			})
			changeAmountCents -= numBills * billValue
		}

	}
	return changeBills, nil
}
