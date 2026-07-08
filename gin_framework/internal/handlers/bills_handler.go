package handlers

import (
	"github.com/carloscfgos1980/cashier-app/internal/config"
	"github.com/gin-gonic/gin"
)

// Define a struct to format the response
type BillRequest struct {
	ID           string  `json:"id"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	Denomination float32 `json:"denomination"`
	Quantity     int32   `json:"quantity"`
}

func GetBillsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use cfg.DB to access the database and retrieve bills
		bills, err := cfg.DB.GetBills(c)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve bills"})
			return
		}
		// Create a slice to hold the formatted bills for the response
		var response []BillRequest
		for _, b := range bills {
			response = append(response, BillRequest{
				ID:           b.ID.String(),
				CreatedAt:    b.CreatedAt.String(),
				UpdatedAt:    b.UpdatedAt.String(),
				Denomination: float32(b.Denomination) / 100, // Convert cents to euros
				Quantity:     b.Quantity,
			})
		}

		c.JSON(200, gin.H{"bills": response})
	}
}
