package sqlc

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferParams) (TransferTxResult, error)
	CreateUserTx(ctx context.Context, arg UserParams) (UserTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

// this way like inheritance in java to make the queries can use transaction
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// create new queries object with transaction ,the callback fucntion and parameter deceide whether commit or rollback base on the err it return
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q) // the function in parameter
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v ,rb err: %v", err, rbErr)
		}
		return err

	}
	return tx.Commit()
}
