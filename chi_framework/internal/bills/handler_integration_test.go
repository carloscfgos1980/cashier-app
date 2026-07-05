package bills

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carloscfgos1980/cashier-app/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBillsService struct {
	billsByDenomination map[int32]database.Bill
	idToDenomination    map[string]int32
}

func newMockBillsService(bills []database.Bill) *mockBillsService {
	service := &mockBillsService{
		billsByDenomination: make(map[int32]database.Bill, len(bills)),
		idToDenomination:    make(map[string]int32, len(bills)),
	}

	for _, bill := range bills {
		service.billsByDenomination[bill.Denomination] = bill
		service.idToDenomination[string(bill.ID.Bytes[:])] = bill.Denomination
	}

	return service
}

func (m *mockBillsService) GetBills(ctx context.Context) ([]database.Bill, error) {
	result := make([]database.Bill, 0, len(m.billsByDenomination))
	for _, bill := range m.billsByDenomination {
		result = append(result, bill)
	}
	return result, nil
}

func (m *mockBillsService) GetBillByDenomination(ctx context.Context, denomination int32) (database.Bill, error) {
	bill, ok := m.billsByDenomination[denomination]
	if !ok {
		return database.Bill{}, errors.New("no rows in result set")
	}
	return bill, nil
}

func (m *mockBillsService) CreateBill(ctx context.Context, denomination int32, quantity int32) (database.Bill, error) {
	return database.Bill{}, errors.New("not implemented")
}

func (m *mockBillsService) UpdateBill(ctx context.Context, id pgtype.UUID, quantity int32) (database.Bill, error) {
	denomination, ok := m.idToDenomination[string(id.Bytes[:])]
	if !ok {
		return database.Bill{}, errors.New("bill id not found")
	}

	bill := m.billsByDenomination[denomination]
	bill.Quantity = quantity
	m.billsByDenomination[denomination] = bill
	return bill, nil
}

func (m *mockBillsService) GetBillsTotalAmount(ctx context.Context) (int64, error) {
	var total int64
	for _, bill := range m.billsByDenomination {
		total += int64(bill.Denomination) * int64(bill.Quantity)
	}
	return total, nil
}

func TestGetChangeRoute_Integration_Success(t *testing.T) {
	service := newMockBillsService([]database.Bill{
		{ID: testUUID(1), Denomination: 500, Quantity: 2},
		{ID: testUUID(2), Denomination: 100, Quantity: 3},
	})

	handler := NewHandler(service)
	router := chi.NewRouter()
	router.Post("/api/change", handler.GetChange)

	body := bytes.NewBufferString(`{"amount_due":10,"amount_paid":15}`)
	req := httptest.NewRequest(http.MethodPost, "/api/change", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equalf(t, http.StatusOK, rec.Code, "unexpected response body: %s", rec.Body.String())

	var got []ChangeLine
	err := json.Unmarshal(rec.Body.Bytes(), &got)
	require.NoError(t, err, "failed to decode response")

	require.Len(t, got, 1)

	expectedLine := "1 x €5 = €5"
	assert.Equal(t, expectedLine, got[0].Text)

	updatedBill, err := service.GetBillByDenomination(context.Background(), 500)
	require.NoError(t, err, "failed to get updated bill")

	assert.EqualValues(t, 1, updatedBill.Quantity)
}

func testUUID(seed byte) pgtype.UUID {
	var id pgtype.UUID
	id.Valid = true
	for i := range id.Bytes {
		id.Bytes[i] = seed
	}
	return id
}
