package handlers

import (
	"net/http"

	"github.com/carloscfgos1980/cashier-app/internal/config"
	"github.com/carloscfgos1980/cashier-app/internal/database"
	"github.com/gin-gonic/gin"
)

// Define a struct to format the response
type BillResponse struct {
	ID           string  `json:"id"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	Denomination float32 `json:"denomination"`
	Quantity     int32   `json:"quantity"`
}

type BillRequest struct {
	Denomination float32 `json:"denomination"`
	Quantity     int32   `json:"quantity"`
}

func GetBillsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use cfg.DB to access the database and retrieve bills
		bills, err := cfg.DB.GetBills(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bills"})
			return
		}
		// Create a slice to hold the formatted bills for the response
		var response []BillResponse
		for _, b := range bills {
			response = append(response, BillResponse{
				ID:           b.ID.String(),
				CreatedAt:    b.CreatedAt.String(),
				UpdatedAt:    b.UpdatedAt.String(),
				Denomination: float32(b.Denomination) / 100, // Convert cents to euros
				Quantity:     b.Quantity,
			})
		}

		c.JSON(http.StatusOK, gin.H{"bills": response})
	}
}

func BillsCreateUpdateHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var bills []BillRequest
		if err := c.ShouldBindJSON(&bills); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		for _, b := range bills {
			demon := b.Denomination
			// Validate the denomination
			if !ValidateDenomination(demon) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid denomination"})
				return
			}
			// convert the denomination from float32 to int (cents) to avoid floating point precision issues
			demonCents := int32(demon * 100)
			dbBill, err := cfg.DB.GetBillByDenomination(c, demonCents)
			if err != nil {
				if err.Error() == "sql: no rows in result set" {
					// Bill does not exist, create a new one
					_, err := cfg.DB.CreateBill(c, database.CreateBillParams{
						Denomination: demonCents,
						Quantity:     b.Quantity,
					})
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bill"})
						return
					}
				}
			}
			// check if the bill already exists
			if dbBill.ID.String() != "" {
				// Bill exists, update the quantity
				_, err := cfg.DB.UpdateBill(c, database.UpdateBillParams{
					ID:       dbBill.ID,
					Quantity: dbBill.Quantity - b.Quantity,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bill"})
					return
				}
			}

		}

		c.JSON(http.StatusOK, bills)
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
