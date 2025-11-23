package pgxpool

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgTransactionFactory struct {
	pool *pgxpool.Pool
}

func NewPgTransactionFactory(pool *pgxpool.Pool) *PgTransactionFactory {
	return &PgTransactionFactory{pool: pool}
}

func (f *PgTransactionFactory) Begin(ctx context.Context) (pgx.Tx, error) {
	return f.pool.Begin(ctx)
}

func (f *PgTransactionFactory) Transaction(ctx context.Context) Transaction {
	tx, ok := ctx.Value(txKey).(Transaction)
	if !ok {
		return f.pool
	}
	return tx
}
