package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/vietquan-37/simplebank/db/mock"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/token"
	"github.com/vietquan-37/simplebank/util"
)

func TestGetAcountApi(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAcount(user.Username)
	testCase := []struct {
		name          string
		acountId      int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorded *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			acountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, authorizationTypeBearer, user.Username, user.Role, time.Minute, tokenMaker)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:     "NotFound",
			acountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, authorizationTypeBearer, account.Owner, user.Role, time.Minute, tokenMaker)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(sqlc.Account{}, sqlc.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)

			},
		},
		{
			name:     "InternalError",
			acountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, authorizationTypeBearer, account.Owner, user.Role, time.Minute, tokenMaker)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(sqlc.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

			},
		},
		{
			name:     "InvalidID",
			acountId: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, authorizationTypeBearer, account.Owner, user.Role, time.Minute, tokenMaker)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name:     "UnAuthorization",
			acountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, authorizationTypeBearer, "unauthorized_user", user.Role, time.Minute, tokenMaker)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
	}
	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl) //this was like the fake db
			tc.buildStubs(store)
			//start test server send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/account/%d", tc.acountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})

	}
}
func randomAcount(owner string) sqlc.Account {
	return sqlc.Account{
		ID:       util.RandomInit(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account sqlc.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotAccount sqlc.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
