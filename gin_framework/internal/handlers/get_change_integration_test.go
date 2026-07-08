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
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DB_URL")
	}
	if dbURL == "" {
		t.Skip("set TEST_DB_URL (or DB_URL) to run integration tests")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dbConn.Close())
	})

	require.NoError(t, dbConn.Ping())

	ctx := context.Background()
	queries := database.New(dbConn)

	_, err = dbConn.ExecContext(ctx, "TRUNCATE TABLE bills")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := dbConn.ExecContext(context.Background(), "TRUNCATE TABLE bills")
		require.NoError(t, cleanupErr)
	})

	_, err = queries.CreateBill(ctx, database.CreateBillParams{Denomination: 500, Quantity: 1})
	require.NoError(t, err)
	_, err = queries.CreateBill(ctx, database.CreateBillParams{Denomination: 200, Quantity: 2})
	require.NoError(t, err)
	_, err = queries.CreateBill(ctx, database.CreateBillParams{Denomination: 100, Quantity: 5})
	require.NoError(t, err)

	cfg := &config.Config{DB: queries}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/change", GetChangeHandler(cfg))

	body, err := json.Marshal(gin.H{
		"amount_due":  13,
		"amount_paid": 20,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/change", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var changeResponse []ChangeLine
	err = json.Unmarshal(w.Body.Bytes(), &changeResponse)
	require.NoError(t, err)
	require.Len(t, changeResponse, 2)

	bill5, err := queries.GetBillByDenomination(ctx, 500)
	require.NoError(t, err)
	require.Equal(t, int32(0), bill5.Quantity)

	bill2, err := queries.GetBillByDenomination(ctx, 200)
	require.NoError(t, err)
	require.Equal(t, int32(1), bill2.Quantity)
}
