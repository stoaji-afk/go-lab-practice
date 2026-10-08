package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/yourname/finops-service/internal/models"
	"github.com/yourname/finops-service/internal/repositories"
)

type TransactionService struct {
	txRepo   repositories.TransactionRepository
	userRepo repositories.UserRepository
	pool     *pgxpool.Pool
}

func NewTransactionService(
	txRepo repositories.TransactionRepository,
	userRepo repositories.UserRepository,
	pool *pgxpool.Pool,
) *TransactionService {
	return &TransactionService{
		txRepo:   txRepo,
		userRepo: userRepo,
		pool:     pool,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tx *models.Transaction) (int, error) {
	if !tx.Amount.GreaterThan(decimal.Zero) {
		return 0, nil
	}

	if tx.Type != "deposit" && tx.Type != "withdraw" {
		return 0, nil
	}

	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer dbTx.Rollback(ctx)

	var userID int
	err = s.pool.QueryRow(ctx, "SELECT id FROM users WHERE id = $1", tx.UserID).Scan(&userID)
	if err != nil {
		return 0, err
	}

	if tx.Type == "withdraw" {
		balance, err := s.userRepo.GetBalance(ctx, tx.UserID)
		if err != nil {
			return 0, err
		}
		if balance.LessThan(tx.Amount) {
			return 0, nil
		}
	}

	newBalance, err := s.userRepo.GetBalance(ctx, tx.UserID)
	if err != nil {
		return 0, err
	}

	if tx.Type == "deposit" {
		newBalance = newBalance.Add(tx.Amount)
	} else {
		newBalance = newBalance.Sub(tx.Amount)
	}

	err = s.userRepo.UpdateBalance(ctx, tx.UserID, newBalance, dbTx)
	if err != nil {
		return 0, err
	}

	tx.Processed = false
	id, err := s.txRepo.CreateTransaction(ctx, tx, dbTx)
	if err != nil {
		return 0, err
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, id int) (*models.Transaction, error) {
	return s.txRepo.GetTransaction(ctx, id)
}

func (s *TransactionService) UpdateTransaction(ctx context.Context, id int, newTx *models.Transaction) error {
	existing, err := s.txRepo.GetTransaction(ctx, id)
	if err != nil {
		return err
	}

	if existing.Processed {
		return nil
	}

	return s.txRepo.UpdateTransaction(ctx, id, newTx)
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, id int) error {
	tx, err := s.txRepo.GetTransaction(ctx, id)
	if err != nil {
		return err
	}

	if tx.Processed {
		balance, err := s.userRepo.GetBalance(ctx, tx.UserID)
		if err != nil {
			return err
		}

		if tx.Type == "deposit" {
			balance = balance.Sub(tx.Amount)
		} else {
			balance = balance.Add(tx.Amount)
		}

		err = s.userRepo.UpdateBalance(ctx, tx.UserID, balance, nil)
		if err != nil {
			return err
		}
	}

	return s.txRepo.DeleteTransaction(ctx, id)
}
