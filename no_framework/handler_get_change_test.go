package main

import (
	"testing"

	"github.com/carloscfgos1980/cashier-app/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateChangeUsesAvailableBills(t *testing.T) {
	bills := []database.Bill{
		{Denomination: 2000, Quantity: 1},
		{Denomination: 1000, Quantity: 2},
		{Denomination: 500, Quantity: 3},
	}

	got, err := calculateChange(3500, bills)
	require.NoError(t, err)

	want := []ChangeBill{
		{Denomination: 20, Quantity: 1},
		{Denomination: 10, Quantity: 1},
		{Denomination: 5, Quantity: 1},
	}

	assert.Equal(t, want, got)
}

func TestFormatChangeResponseBuildsDisplayLines(t *testing.T) {
	changeBills := []ChangeBill{
		{Denomination: 20, Quantity: 2},
		{Denomination: 0.5, Quantity: 1},
	}

	got := formatChangeResponse(changeBills)
	want := []ChangeLine{
		{Text: "2 x €20 = 40"},
		{Text: "1 x €0.5 = 0.5"},
	}

	assert.Equal(t, want, got)
}
