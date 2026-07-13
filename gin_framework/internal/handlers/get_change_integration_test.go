package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/carloscfgos1980/cashier-app/internal/config"
	"github.com/carloscfgos1980/cashier-app/internal/database"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestGetChangeRoute_Integration(t *testing.T) {
	// Check for the TEST_DB_URL environment variable, fallback to DB_URL if not set.
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DB_URL")
	}
	if dbURL == "" {
		t.Skip("set TEST_DB_URL (or DB_URL) to run integration tests")
	}
	// Connect to the database.
	dbConn, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dbConn.Close())
	})

	require.NoError(t, dbConn.Ping())
	// Create a new database queries instance.
	ctx := context.Background()
	queries := database.New(dbConn)
	// Clear the bills table before running the test to ensure a clean state.
	_, err = dbConn.ExecContext(ctx, "TRUNCATE TABLE bills")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := dbConn.ExecContext(context.Background(), "TRUNCATE TABLE bills")
		require.NoError(t, cleanupErr)
	})
	// Insert test bills into the database.
	_, err = queries.CreateBill(ctx, database.CreateBillParams{Denomination: 500, Quantity: 1})
	require.NoError(t, err)
	_, err = queries.CreateBill(ctx, database.CreateBillParams{Denomination: 200, Quantity: 2})
	require.NoError(t, err)
	_, err = queries.CreateBill(ctx, database.CreateBillParams{Denomination: 100, Quantity: 5})
	require.NoError(t, err)
	// Create a config instance with the database queries.
	cfg := &config.Config{DB: queries}
	// Set Gin to test mode and create a new router for testing.
	gin.SetMode(gin.TestMode)
	// Create a new Gin router and register the GetChangeHandler route.
	router := gin.New()
	// Register the GetChangeHandler route with the router.
	router.POST("/api/change", GetChangeHandler(cfg))
	// Prepare the request body for the change request.
	body, err := json.Marshal(gin.H{
		"amount_due":  13,
		"amount_paid": 20,
	})
	require.NoError(t, err)
	// Create a new HTTP request to the /api/change endpoint with the prepared body.
	req := httptest.NewRequest(http.MethodPost, "/api/change", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	// Serve the HTTP request using the router and record the response.
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	// Unmarshal the response body into a slice of ChangeLine structs.
	var changeResponse []ChangeLine
	err = json.Unmarshal(w.Body.Bytes(), &changeResponse)
	require.NoError(t, err)
	require.Len(t, changeResponse, 2)
	// Check the first change line for the 5 euro bill.
	bill5, err := queries.GetBillByDenomination(ctx, 500)
	require.NoError(t, err)
	require.Equal(t, int32(0), bill5.Quantity)
	// Check the second change line for the 2 euro bill.
	bill2, err := queries.GetBillByDenomination(ctx, 200)
	require.NoError(t, err)
	require.Equal(t, int32(1), bill2.Quantity)
}
