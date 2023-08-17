package txpool

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type transactionCtxKeyType struct{}

var transactionContextKey = transactionCtxKeyType{}

// CtxWithTransaction puts a transaction into context. Can be used with
// individual methods of TxPool instead of TxPool.Transaction to avoid nesting.
func CtxWithTransaction(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, transactionContextKey, tx)
}
