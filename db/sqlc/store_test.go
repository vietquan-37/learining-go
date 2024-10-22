package sqlc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := RandomAccount(t)
	account2 := RandomAccount(t)
	//run n concurrent transfer transaction
	n := 5
	amount := int64(10)
	errs := make(chan error)
	results := make(chan TransferTxResult)
	for i := 0; i < n; i++ {
		//we can not check right here because go routine run in a seperate from main so use channel
		go func() {
			result, err := store.TransferTx(context.Background(), TransferParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err //sending error to errrs channel
			results <- result
		}()
	}
	for i := 0; i < n; i++ {
		err := <-errs //receive
		require.NoError(t, err)
		result := <-results
		require.NotEmpty(t, result)
		//check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)
		//check from entries
		FromEntry := result.FromEntry
		require.NotEmpty(t, FromEntry)
		require.Equal(t, account1.ID, FromEntry.AccountID)
		require.Equal(t, -amount, FromEntry.Amount)
		require.NotZero(t, FromEntry.ID)
		require.NotZero(t, FromEntry.CreatedAt)
		_, err = store.GetOneEntry(context.Background(), FromEntry.ID)
		require.NoError(t, err)
		//to entry check
		ToEntry := result.ToEntry
		require.NotEmpty(t, ToEntry)
		require.Equal(t, account2.ID, ToEntry.AccountID)
		require.Equal(t, amount, ToEntry.Amount)
		require.NotZero(t, ToEntry.ID)
		require.NotZero(t, ToEntry.CreatedAt)
		_, err = store.GetOneEntry(context.Background(), ToEntry.ID)
		require.NoError(t, err)
		//check account balance
	}
}
