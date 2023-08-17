# txpool

txpool is a simple wrapper for frequently used pgxpool methods (from [pgx](https://github.com/jackc/pgx/tree/master)
library) to allow adding an optional transaction.

Before executing queries, it checks `context.Context` for a `pgx.Tx`, and runs the query on it, if its present.

## Motivation

A typical repository would look like this, wrapping an instance of `pgxpool.Pool`:

```go
type Repository struct {
    pool *pgxpool.Pool
}

func (r *Repository) CountRows(ctx context.Context) (int, error) {
    row := r.pgx.QueryRow(ctx, "select count(*) from some_table")
    // ...
}

func (r *Repository) DoStuff(ctx context.Context) error {
    row := r.pgx.QueryRow(ctx, "insert (id) values (1) into some_table")
    // ...
}
```

To execute these two queries in a single transaction, you would need to add an optional `pgx.Tx` argument,
add nil checks, etc; alternatively you could create copies of those methods, thus unnecessarily duplicating code.

Using txpool, you can simply do this:

```go
type Repository struct {
    pool *txpool.Pool // pgxpool -> txpool
}

func (r *Repository) CountRows(ctx context.Context) (int, error) {
    row := r.pool.QueryRow(ctx, "select count(*) from some_table")
    // ...
}

func (r *Repository) DoStuff(ctx context.Context) error {
    row := r.pool.QueryRow(ctx, "insert (id) values (1) into some_table")
    // ...
}
```

And use them like this:

```go
func (r *PgxRepository) CallBoth(ctx context.Context) error {
	return r.pool.Transaction(ctx, func(ctx context.Context) error {
		count, err := r.pool.CountRows(ctx)
		// ...
		err = r.pool.DoStuff(ctx)
		// ...
		return nil
	})
}
```

## Accessing other pgxpool.Pool methods

Use this method to access the underlying `pgxpool.Pool`:

```go
func (t *TxPool) Pool() *pgxpool.Pool {
	return t.pool
}
```

## Note

txpool currently only supports a few general methods of pgxpool