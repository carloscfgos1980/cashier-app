package bills

import (
	"context"

	"github.com/carloscfgos1980/cashier-app/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service defines the interface for the users service
type Service interface {
	GetBills(ctx context.Context) ([]database.Bill, error)
	GetBillByDenomination(ctx context.Context, denomination int32) (database.Bill, error)
	CreateBill(ctx context.Context, denomination int32, quantity int32) (database.Bill, error)
	UpdateBill(ctx context.Context, id pgtype.UUID, quantity int32) (database.Bill, error)
	GetBillsTotalAmount(ctx context.Context) (int64, error)
}

// svc defines the struct for the users service
type svc struct {
	repo *database.Queries
	db   *pgx.Conn
}

// NewService creates a new service for the users package
func NewService(repo *database.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

// GetBills retrieves all bills from the database
func (s *svc) GetBills(ctx context.Context) ([]database.Bill, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// create a new Queries instance with the transaction
	repo := database.New(tx)

	// get the bills from the repository
	bills, err := repo.GetBills(ctx)
	if err != nil {
		return nil, err
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return bills, nil
}

// GetBillByDenomination retrieves a bill by its denomination from the database
func (s *svc) GetBillByDenomination(ctx context.Context, denomination int32) (database.Bill, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.Bill{}, err
	}
	defer tx.Rollback(ctx)

	// create a new Queries instance with the transaction
	repo := database.New(tx)

	// get the bill from the repository
	bill, err := repo.GetBillByDenomination(ctx, denomination)
	if err != nil {
		return database.Bill{}, err
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.Bill{}, err
	}

	return bill, nil
}

// CreateBill creates a new bill in the database
func (s *svc) CreateBill(ctx context.Context, denomination int32, quantity int32) (database.Bill, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.Bill{}, err
	}
	defer tx.Rollback(ctx)

	// create a new Queries instance with the transaction
	repo := database.New(tx)

	// create the bill in the repository
	params := database.CreateBillParams{
		Denomination: denomination,
		Quantity:     quantity,
	}
	bill, err := repo.CreateBill(ctx, params)
	if err != nil {
		return database.Bill{}, err
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.Bill{}, err
	}

	return bill, nil
}

// UpdateBill updates the quantity of a bill in the database
func (s *svc) UpdateBill(ctx context.Context, id pgtype.UUID, quantity int32) (database.Bill, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.Bill{}, err
	}
	defer tx.Rollback(ctx)

	// create a new Queries instance with the transaction
	repo := database.New(tx)

	// update the bill in the repository
	params := database.UpdateBillParams{
		ID:       id,
		Quantity: quantity,
	}
	bill, err := repo.UpdateBill(ctx, params)
	if err != nil {
		return database.Bill{}, err
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.Bill{}, err
	}

	return bill, nil
}

// GetBillsTotalAmount retrieves the total amount of all bills in the database
func (s *svc) GetBillsTotalAmount(ctx context.Context) (int64, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// create a new Queries instance with the transaction
	repo := database.New(tx)

	// get the total amount from the repository
	totalAmount, err := repo.GetBillsTotalAmount(ctx)
	if err != nil {
		return 0, err
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return totalAmount, nil
}
