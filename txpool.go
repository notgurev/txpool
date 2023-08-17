package txpool

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxPool is a wrapper for pgxpool.Pool which can accept optional transactions
// through context.Context.
type TxPool struct {
	pool *pgxpool.Pool
}

var _ Querier = (*TxPool)(nil)

func New(pool *pgxpool.Pool) *TxPool {
	return &TxPool{pool: pool}
}

// Transaction wraps the function in a transaction, handles rollback and commit.
func (t *TxPool) Transaction(ctx context.Context, f func(ctx context.Context) error) (err error) {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		er := tx.Rollback(ctx)
		if er != nil {
			if errors.Is(er, pgx.ErrTxClosed) {
				return
			}
			err = errors.Join(err, er)
		}
	}()

	ctx = context.WithValue(ctx, transactionContextKey, tx)

	if err = f(ctx); err != nil {
		return fmt.Errorf("run wrapped function in tx: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return
}

// Querier returns a transaction if it is present in context. Otherwise, returns
// the underlying pgxpool.Pool.
func (t *TxPool) Querier(ctx context.Context) Querier {
	tx, ok := ctx.Value(transactionContextKey).(Querier)
	if !ok {
		return t.pool
	}

	return tx
}

func (t *TxPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tag, err := t.Querier(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("txpool exec: %w", err)
	}

	return tag, nil
}

func (t *TxPool) Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error) {
	rows, err := t.Querier(ctx).Query(ctx, sql, optionsAndArgs...)
	if err != nil {
		return nil, fmt.Errorf("txpool query: %w", err)
	}

	return rows, nil
}

func (t *TxPool) QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row {
	return t.Querier(ctx).QueryRow(ctx, sql, optionsAndArgs...)
}

// Pool returns the underlying pgxpool.Pool.
func (t *TxPool) Pool() *pgxpool.Pool {
	return t.pool
}
