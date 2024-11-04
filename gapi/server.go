package gapi

import (
	"fmt"

	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/pb"
	"github.com/vietquan-37/simplebank/token"
	"github.com/vietquan-37/simplebank/util"
	"github.com/vietquan-37/simplebank/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config          util.Config
	store           sqlc.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store sqlc.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil

}
