package sqlc

import (
	"context"
)

type UserParams struct {
	CreateUserParams
	AfterCreate func(user User) error
}

type UserTxResult struct {
	User User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg UserParams) (UserTxResult, error) {
	var result UserTxResult
	err := store.execTx(ctx, func(q *Queries) error { // this become a clousure function when we want to get to result from call back fucntion
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}
		return arg.AfterCreate(result.User)
	})
	return result, err
}
