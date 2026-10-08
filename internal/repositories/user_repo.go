package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetBalance(ctx context.Context, userID int) (decimal.Decimal, error) {
	var balance decimal.Decimal
	err := r.pool.QueryRow(ctx, "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func (r *UserRepo) UpdateBalance(ctx context.Context, userID int, amount decimal.Decimal, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, "UPDATE users SET balance = $1 WHERE id = $2", amount, userID)
	if err != nil {
		return err
	}
	return nil
}
