package gapi

import (
	"context"

	"github.com/lib/pq"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/pb"
	"github.com/vietquan-37/simplebank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	hashPassword, err := util.HashedPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to hashed password: %s", err)
	}
	arg := sqlc.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashPassword,
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "email has been register before: %s", err)

			}
		}
		return nil, status.Errorf(codes.Internal, "fail to register user: %s", err)
	}
	res := &pb.CreateUserResponse{
		User: convertUser(user),
	}
	return res, nil

}
