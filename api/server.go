package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/vietquan-37/simplebank/db/sqlc"
)

type Server struct {
	store  sqlc.Store
	router *gin.Engine
}

func NewServer(store sqlc.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}
	router.POST("/account", server.createAccount)
	router.POST("/register", server.createUser)
	router.POST("/transfer", server.createTransfer)
	router.GET("/account/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	server.router = router
	return server

}

// this run http server on specific address
func (sever *Server) Start(address string) error {
	return sever.router.Run(address)
}

// gin.H was a map string value
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
