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
	var response []Bill
	for _, bill := range bills {
		response = append(response, Bill{
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
