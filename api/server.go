package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/token"
	"github.com/vietquan-37/simplebank/util"
)

type Server struct {
	config     util.Config
	store      sqlc.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store sqlc.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setUpRouter()
	return server, nil

}
func (server *Server) setUpRouter() {
	router := gin.Default()

	router.POST("/register", server.createUser)
	router.POST("/login", server.login)
	router.POST("refresh", server.newAccessToken)
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/account", server.createAccount)
	authRoutes.POST("/transfer", server.createTransfer)
	authRoutes.GET("/account/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccountByOwner)
	server.router = router

}

// this run http server on specific address
func (sever *Server) Start(address string) error {
	return sever.router.Run(address)
}

// gin.H was a map string value
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
