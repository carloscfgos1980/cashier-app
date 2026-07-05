package bills

import (
	"context"

	"github.com/carloscfgos1980/cashier-app/internal/database"
	"github.com/jackc/pgx/v5"
)

// Service defines the interface for the users service
type Service interface {
	GetBills(ctx context.Context) ([]database.Bill, error)
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
