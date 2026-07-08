package handlers

import (
	"net/http"

	"github.com/carloscfgos1980/cashier-app/internal/config"
	"github.com/gin-gonic/gin"
)

type MetricsResponse struct {
	DenominationsCount int32   `json:"denominations_count"`
	TotalBillsCount    int32   `json:"total_bills_count"`
	TotalAmountCents   int64   `json:"total_amount_cents"`
	TotalAmountEuro    float64 `json:"total_amount_euro"`
}

func GetMetricsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		bills, err := cfg.DB.GetBills(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bills"})
			return
		}

		totalAmountCents, err := cfg.DB.GetBillsTotalAmount(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve total amount"})
			return
		}

		var totalBillsCount int32
		for _, bill := range bills {
			totalBillsCount += bill.Quantity
		}

		response := MetricsResponse{
			DenominationsCount: int32(len(bills)),
			TotalBillsCount:    totalBillsCount,
			TotalAmountCents:   totalAmountCents,
			TotalAmountEuro:    float64(totalAmountCents) / 100,
		}

		c.JSON(http.StatusOK, response)
	}
}
