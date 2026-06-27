package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
)

func main() {
	amountDue, amountPaid, err := parseAmountsFromCLI(os.Args)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Usage: go run ./3.cashier <amountDue> <amountPaid>")
		fmt.Println("Example: go run ./3.cashier 13.37 100.00")
		return
	}

	// Cashier's available denominations and their counts
	cashier := map[float64]int{
		100.00: 10,  // $100 bills
		50.00:  0,   // $50 bills
		20.00:  30,  // $20 bills
		10.00:  40,  // $10 bills
		5.00:   50,  // $5 bills
		1.00:   100, // $1 bills
		0.25:   200, // Quarters
		0.10:   300, // Dimes
		0.05:   400, // Nickels
		0.01:   500, // Pennies
	}

	// Calculate the change to be given
	change, err := CalculateChange(amountDue, amountPaid, cashier)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		// Print Change calculation summary
		totalChange := amountPaid - amountDue
		fmt.Println("Change calculation summary")
		fmt.Printf("Amount due:   $%.2f\n", amountDue)
		fmt.Printf("Amount paid:  $%.2f\n", amountPaid)
		fmt.Printf("Total change: $%.2f\n\n", totalChange)

		// Separate bills and coins for better readability
		var bills []float64
		var coins []float64
		for denom := range change {
			if denom >= 1.00 {
				bills = append(bills, denom)
			} else {
				coins = append(coins, denom)
			}
		}

		// Sort bills and coins in descending order for display
		sort.Sort(sort.Reverse(sort.Float64Slice(bills)))
		sort.Sort(sort.Reverse(sort.Float64Slice(coins)))

		billsSubtotal := 0.0
		coinsSubtotal := 0.0

		// Print the change distribution
		fmt.Println("Bills:")
		if len(bills) == 0 {
			fmt.Println("  none")
		} else {
			for _, denom := range bills {
				count := change[denom]
				lineTotal := denom * float64(count)
				billsSubtotal += lineTotal
				fmt.Printf("  $%6.2f x %-3d = $%6.2f\n", denom, count, lineTotal)
			}
		}

		fmt.Println("\nCoins:")
		if len(coins) == 0 {
			fmt.Println("  none")
		} else {
			for _, denom := range coins {
				count := change[denom]
				lineTotal := denom * float64(count)
				coinsSubtotal += lineTotal
				fmt.Printf("  $%6.2f x %-3d = $%6.2f\n", denom, count, lineTotal)
			}
		}

		// Print subtotals and grand total
		fmt.Printf("\nBills subtotal: $%.2f\n", billsSubtotal)
		fmt.Printf("Coins subtotal: $%.2f\n", coinsSubtotal)
		fmt.Printf("Grand total:    $%.2f\n", billsSubtotal+coinsSubtotal)

		// Print the remaining cashier inventory after the transaction
		var cashierBills []float64
		var cashierCoins []float64
		for denom := range cashier {
			if denom >= 1.00 {
				cashierBills = append(cashierBills, denom)
			} else {
				cashierCoins = append(cashierCoins, denom)
			}
		}

		sort.Sort(sort.Reverse(sort.Float64Slice(cashierBills)))
		sort.Sort(sort.Reverse(sort.Float64Slice(cashierCoins)))

		fmt.Println("\nCashier inventory after payment")
		fmt.Println("Bills:")
		for _, denom := range cashierBills {
			fmt.Printf("  $%6.2f -> %d\n", denom, cashier[denom])
		}

		fmt.Println("\nCoins:")
		for _, denom := range cashierCoins {
			fmt.Printf("  $%6.2f -> %d\n", denom, cashier[denom])
		}
	}
}

func parseAmountsFromCLI(args []string) (float64, float64, error) {
	if len(args) != 3 {
		return 0, 0, fmt.Errorf("expected 2 arguments, got %d", len(args)-1)
	}

	amountDue, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid amountDue: %w", err)
	}

	amountPaid, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid amountPaid: %w", err)
	}

	return amountDue, amountPaid, nil
}

func CalculateChange(amountDue float64, amountPaid float64, cashier map[float64]int) (map[float64]int, error) {
	// Validate input amounts
	if amountDue < 0 || amountPaid < 0 {
		return nil, fmt.Errorf("amounts cannot be negative")
	}
	// Validate that the amount paid is sufficient
	if amountPaid < amountDue {
		return nil, fmt.Errorf("amount paid is less than amount due")
	}
	// Calculate the change in cents
	changeCents := int(math.Round((amountPaid - amountDue) * 100))
	// Create a map to hold the change distribution
	changeDistribution := make(map[float64]int)

	// Convert the cashier map to a slice of denominations in cents for easier processing
	denominations := make([]int, 0, len(cashier))
	// Create a map to convert back from cents to float64 denominations
	denomByCents := make(map[int]float64, len(cashier))
	// Populate the denominations slice and the conversion map
	for denom := range cashier {
		cents := int(math.Round(denom * 100))
		denominations = append(denominations, cents)
		denomByCents[cents] = denom
	}
	// Sort the denominations in descending order to prioritize larger bills and coins
	sort.Sort(sort.Reverse(sort.IntSlice(denominations)))
	// Iterate through the sorted denominations to calculate the change distribution
	for _, denom := range denominations {
		// If the remaining change is zero or less, we can break early
		if changeCents <= 0 {
			break
		}
		log.Printf("Processing denomination: %d cents, Change left: %d cents\n", denom, changeCents)
		// value of the denomination in float64 for logging and map access
		denomValue := denomByCents[denom]
		log.Printf("Cashier has %d of $%.2f\n", cashier[denomValue], denomValue)
		// If the cashier has this denomination and the change left is greater than or equal to the denomination, we can use it
		if cashier[denomValue] > 0 && changeCents >= denom {
			numBillsCoins := changeCents / denom // Calculate the maximum number of this denomination that can be used
			// Ensure we don't use more than what the cashier has
			if numBillsCoins > cashier[denomValue] {
				numBillsCoins = cashier[denomValue]
				log.Printf("Limited by cashier's inventory. Using %d of $%.2f\n", numBillsCoins, denomValue)
			}
			// Update the change distribution and the remaining change
			if numBillsCoins > 0 {
				changeDistribution[denomValue] = numBillsCoins
				changeCents -= numBillsCoins * denom
				cashier[denomValue] -= numBillsCoins
				log.Printf("Using %d of $%.2f, Change left: %d cents\n", numBillsCoins, denomValue, changeCents)
			}
		}
	}
	// If there is still change left to give, it means the cashier doesn't have enough denominations to provide the exact change
	if changeCents > 0 {
		log.Printf("Insufficient change available. Change left: %d cents\n", changeCents)
		return nil, fmt.Errorf("insufficient change available")
	}
	// Return the calculated change distribution
	return changeDistribution, nil
}
