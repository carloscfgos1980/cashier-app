package bills

import (
	"net/http"

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
			Denomination: float32(bill.Denomination),
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

func ValidateDenomination(denomination float32) bool {
	switch denomination {
	case 100.00, 50.00, 20.00, 10.00, 5.00, 1.00, 0.50, 0.20, 0.10, 0.05, 0.01:
		return true
	default:
		return false
	}
}
