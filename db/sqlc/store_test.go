package sqlc

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := RandomAccount(t)
	account2 := RandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)
	//run n concurrent transfer transaction
	n := 5
	amount := int64(10)
	errs := make(chan error)
	results := make(chan TransferTxResult)
	for i := 0; i < n; i++ {

		//we can not check right here because go routine run in a seperate from main so use channel
		go func() {
			// this was like the map as the key is the key in map and txname is the value

			result, err := store.TransferTx(context.Background(), TransferParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err //sending error to errrs channel
			results <- result
		}()
	}
	existed := make(map[int]bool)
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
		FromAccount := result.FromAccount
		require.NotEmpty(t, FromAccount)
		require.Equal(t, account1.ID, FromAccount.ID)

		ToAccount := result.ToAccount
		require.NotEmpty(t, ToAccount)
		require.Equal(t, account2.ID, ToAccount.ID)
		fmt.Println(">> tx:", FromAccount.Balance, ToAccount.Balance)
		//check the account balance
		diff1 := account1.Balance - FromAccount.Balance
		diff2 := ToAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff2 > 0)
		require.True(t, diff1%amount == 0)
		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		// require.False(t, existed[k])
		require.NotContains(t, existed, k)
		existed[k] = true
		//check final update balance

	}

	updateAcount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updateAcount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", account1.Balance, account2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updateAcount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*int64(amount), updateAcount2.Balance)

}

// the deadlock can be occur when the account1 send money to account 2 and account2 send money to account 1 too
func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)
	account1 := RandomAccount(t)
	account2 := RandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)
	//run n concurrent transfer transaction
	n := 10
	amount := int64(10)
	errs := make(chan error)

	for i := 0; i < n; i++ {

		//we can not check right here because go routine run in a seperate from main so use channel
		go func() {
			FromAccountID := account1.ID
			ToAccountID := account2.ID
			if i%2 == 1 {
				FromAccountID = account2.ID
				ToAccountID = account1.ID
			}
			_, err := store.TransferTx(context.Background(), TransferParams{
				FromAccountID: FromAccountID,
				ToAccountID:   ToAccountID,
				Amount:        amount,
			})
			errs <- err //sending error to errrs channel

		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs //receive
		require.NoError(t, err)

	}

	updateAcount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updateAcount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updateAcount1.Balance, updateAcount2.Balance)
	require.Equal(t, account1.Balance, updateAcount1.Balance)
	require.Equal(t, account2.Balance, updateAcount2.Balance)

}
