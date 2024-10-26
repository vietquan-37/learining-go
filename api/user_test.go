package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	mockdb "github.com/vietquan-37/simplebank/db/mock"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/util"
)

type eqCreateUserParamsMatcher struct {
	arg      sqlc.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(sqlc.CreateUserParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}
	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v ", e.arg, e.password)
}
func EqCreateUserParams(arg sqlc.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}
func TestCreateUserApi(t *testing.T) {
	user, password := randomUser(t)
	testCase := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorded *httptest.ResponseRecorder)
	}{
		{
			name: "Created",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullname": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).Times(1).Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InvalidError",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullname": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(1).
					Return(sqlc.User{}, sql.ErrConnDone)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullname": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(1).
					Return(sqlc.User{}, &pq.Error{Code: "23505"})

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)

			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "loancho@1",
				"password": password,
				"fullname": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullname": user.FullName,
				"email":    "loancho",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username": user.Username,
				"password": "123",
				"fullname": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

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
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			//change body data to Json
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)
			url := "/register"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})

	}
}

func randomUser(t *testing.T) (user sqlc.User, password string) {
	password = util.RandomString(6)
	hashPassword, err := util.HashedPassword(password)
	require.NoError(t, err)
	user = sqlc.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return

}
func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user sqlc.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotUser sqlc.User
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}