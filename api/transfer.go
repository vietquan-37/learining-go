package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/token"
)

type CreateTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,min=1"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req CreateTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account not belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	if req.FromAccountID == req.ToAccountID {
		err := errors.New("cannot send money to your own account")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if fromAccount.Balance < req.Amount {
		err := errors.New("dont have enough balance")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	_, valid = server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	arg := sqlc.TransferParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}
	transfer, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusCreated, transfer)
}
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (sqlc.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if errors.Is(err, sqlc.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}
	return account, true
}
