package handlers

import (
	"errors"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

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

// Define a struct to handle the request for creating or updating bills
type BillRequest struct {
	Denomination float32 `json:"denomination"`
	Quantity     int32   `json:"quantity"`
}

// GetBillsHandler retrieves all bills from the database and returns them in a formatted response.
func GetBillsHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that retrieves bills from the database and formats the response.
	return func(c *gin.Context) {
		// Use cfg.DB to access the database and retrieve bills
		bills, err := cfg.DB.GetBills(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bills"})
			return
		}
		// Create a slice to hold the formatted bills for the response
		var response []BillResponse
		// Format each bill and append it to the response slice
		for _, b := range bills {
			response = append(response, BillResponse{
				ID:           b.ID.String(),
				CreatedAt:    b.CreatedAt.String(),
				UpdatedAt:    b.UpdatedAt.String(),
				Denomination: float32(b.Denomination) / 100, // Convert cents to euros
				Quantity:     b.Quantity,
			})
		}
		// Return the formatted bills as a JSON response
		c.JSON(http.StatusOK, response)
	}
}

// BillsCreateUpdateHandler handles the creation or update of bills in the database.
func BillsCreateUpdateHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that processes the request to create or update bills.
	return func(c *gin.Context) {
		// Parse the request body into a slice of BillRequest structs
		var bills []BillRequest
		if err := c.ShouldBindJSON(&bills); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// Iterate over each bill in the request and process it
		for _, b := range bills {
			demon := b.Denomination
			// Validate the quantity
			if b.Quantity < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be greater than or equal to 0"})
				return
			}
			// Validate the denomination
			if !ValidateDenomination(demon) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid denomination"})
				return
			}
			// convert the denomination from float32 to int (cents) to avoid floating point precision issues
			demonCents := int32(demon * 100)
			// Check if the bill already exists in the database
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
					Quantity: b.Quantity,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bill"})
					return
				}
			}

		}
		// Return bills after processing the request
		c.JSON(http.StatusOK, bills)
	}
}

// ValidateDenomination checks if the provided denomination is valid based on predefined denominations.
func ValidateDenomination(denomination float32) bool {
	switch denomination {
	case 100.00, 50.00, 20.00, 10.00, 5.00, 1.00, 0.50, 0.20, 0.10, 0.05, 0.01:
		return true
	default:
		return false
	}
}

// TransactionRequest represents the request body for a transaction.
type TransactionRequest struct {
	AmountDue  float32 `json:"amount_due"`
	AmountPaid float32 `json:"amount_paid"`
}

// ChangeLine represents a line in the change response.
type ChangeLine struct {
	Text string `json:"text"`
}

// GetChangeHandler handles the change calculation and updates the bill inventory accordingly.
func GetChangeHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that processes the request to calculate change and update the bill inventory.
	return func(c *gin.Context) {
		// Parse the request body into a TransactionRequest struct
		var request TransactionRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// Validate that the amount paid is not less than the amount due
		if request.AmountPaid < request.AmountDue {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Amount paid is less than amount due"})
			return
		}
		// Calculate the change amount in cents to avoid floating point precision issues
		changeAmount := request.AmountPaid - request.AmountDue
		changeAmountCents := int64(changeAmount * 100) // Convert to cents
		// Get the total amount of bills in the register to check if there are sufficient funds for change.
		totalAmountIntCents, err := cfg.DB.GetBillsTotalAmount(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve total amount"})
			return
		}
		// Check if there are sufficient funds in the register for change.
		if changeAmountCents > totalAmountIntCents {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds in the register for change"})
			return
		}
		// Get the available bills from the database.
		dbBills, err := cfg.DB.GetBills(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bills"})
			return
		}

		// Calculate the change to be given.
		changeBills, err := calculateChange(int32(changeAmountCents), dbBills)
		if err != nil {
			if err.Error() == "insufficient funds in the register for change" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds in the register for change"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate change"})
			return
		}

		// Persist the dispensed change so the current bill inventory stays in sync.
		for _, bill := range changeBills {
			denominationCents := int32(math.Round(float64(bill.Denomination) * 100))
			// Retrieve the bill from the database to get its current quantity.
			dbBill, err := cfg.DB.GetBillByDenomination(c, denominationCents)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bill inventory"})
				return
			}
			// Calculate the new quantity after dispensing the change.
			newQuantity := dbBill.Quantity - bill.Quantity
			if newQuantity < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds in the register for change"})
				return
			}
			// Update the bill quantity in the database.
			_, err = cfg.DB.UpdateBill(c, database.UpdateBillParams{
				ID:       dbBill.ID,
				Quantity: newQuantity,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bill inventory"})
				return
			}
		}

		// Format the change response.
		response := formatChangeResponse(changeBills)
		// Return the formatted change response as JSON.
		c.JSON(http.StatusOK, response)
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
	// Create a slice to hold the change bills to be returned.
	changeBills := make([]BillResponse, 0)
	// Sort the bills in descending order of denomination to give the largest bills first.
	sortedBills := append([]database.Bill(nil), bills...)
	sort.Slice(sortedBills, func(i, j int) bool {
		return sortedBills[i].Denomination > sortedBills[j].Denomination
	})
	// Iterate over the sorted bills and calculate how many of each denomination can be given as change.
	for _, bill := range sortedBills {
		// If the change amount is zero or less, break out of the loop.
		if changeAmountCents <= 0 {
			break
		}
		billValue := bill.Denomination
		billQuantity := bill.Quantity
		// If the bill value is less than or equal to the change amount and there are bills available, calculate how many can be given.
		if billValue <= changeAmountCents && billQuantity > 0 {
			numBills := changeAmountCents / billValue
			if numBills > billQuantity {
				numBills = billQuantity
			}
			// Append the calculated number of bills to the changeBills slice.
			changeBills = append(changeBills, BillResponse{
				Denomination: float32(billValue) / 100,
				Quantity:     numBills,
			})
			// Subtract the value of the given bills from the change amount.
			changeAmountCents -= numBills * billValue
		}

	}
	// If there is still change left to be given, return an error indicating insufficient funds.
	if changeAmountCents > 0 {
		return nil, errors.New("insufficient funds in the register for change")
	}
	// Return the calculated change bills.
	return changeBills, nil
}
