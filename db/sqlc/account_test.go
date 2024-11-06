package sqlc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vietquan-37/simplebank/util"
)

func RandomAccount(t *testing.T) Account {
	user := RandomUser(t)
	arg := CreateAccountParams{
		Owner:    user.Username, //randomly generate data
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := testStore.CreateAccount(context.Background(), arg)
	require.NoError(t, err) // this say that if the error not nil the test will fail
	require.NotEmpty(t, account)
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
	return account
}
func TestCreateAccount(t *testing.T) {
	RandomAccount(t)
}
func TestGetAccount(t *testing.T) {
	//create account
	account1 := RandomAccount(t)
	account2, err := testStore.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)

}
func TestUpdateAccount(t *testing.T) {
	account1 := RandomAccount(t)
	arg := UpdateAccountParams{
		ID:      account1.ID,
		Balance: account1.Balance,
	}
	account2, err := testStore.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account2)
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance) // this should change
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}
func TestDeleteAcount(t *testing.T) {
	account1 := RandomAccount(t)
	err := testStore.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	account2, err := testStore.GetAccount(context.Background(), account1.ID)
	require.Error(t, err)
	require.EqualError(t, err, ErrRecordNoFound.Error())
	require.Empty(t, account2)
}
func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		RandomAccount(t)
	}
	arg := ListAccountParams{
		Limit:  5,
		Offset: 5,
	}
	accounts, err := testStore.ListAccount(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)
	for _, account := range accounts {
		require.NotEmpty(t, account)
	}

}
func TestListAccountsByOwner(t *testing.T) {
	var LastAccount Account
	for i := 0; i < 10; i++ {
		LastAccount = RandomAccount(t)
	}
	arg := GetAccountsByOwnerParams{
		Owner:  LastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}
	accounts, err := testStore.GetAccountsByOwner(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)
	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, LastAccount.Owner, account.Owner)
	}

}
