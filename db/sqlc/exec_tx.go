package sqlc

import (
	"context"
	"fmt"
)

// create new queries object with transaction ,the callback fucntion and parameter deceide whether commit or rollback base on the err it return
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q) // the function in parameter
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v ,rb err: %v", err, rbErr)
		}
		return err

	}
	return tx.Commit(ctx)
}
