package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourname/finops-service/internal/models"
)

type TransactionRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{pool: pool}
}

func (r *TransactionRepo) CreateTransaction(ctx context.Context, tx *models.Transaction, dbTx pgx.Tx) (int, error) {
	var id int
	err := dbTx.QueryRow(
		ctx,
		`INSERT INTO transactions (user_id, amount, type, timestamp, processed)
		 VALUES ($1, $2, $3, NOW(), false)
		 RETURNING id`,
		tx.UserID, tx.Amount.String(), tx.Type,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *TransactionRepo) GetTransaction(ctx context.Context, id int) (*models.Transaction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, amount, type, timestamp, processed
		 FROM transactions WHERE id = $1`, id)

	var tx models.Transaction
	err := row.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.Timestamp, &tx.Processed)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (r *TransactionRepo) UpdateTransaction(ctx context.Context, id int, newTx *models.Transaction) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE transactions SET type = $1, timestamp = $2, processed = $3
		 WHERE id = $4`,
		newTx.Type, newTx.Timestamp, newTx.Processed, id,
	)
	return err
}

func (r *TransactionRepo) DeleteTransaction(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM transactions WHERE id = $1", id)
	return err
}
