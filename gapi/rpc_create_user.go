package gapi

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/pb"
	"github.com/vietquan-37/simplebank/util"
	"github.com/vietquan-37/simplebank/val"
	"github.com/vietquan-37/simplebank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violation := validateCreateUserRequest(req)
	if violation != nil {
		return nil, invalidArgumentError(violation)
	}
	hashPassword, err := util.HashedPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to hashed password: %s", err)
	}
	arg := sqlc.UserParams{
		CreateUserParams: sqlc.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hashPassword,
			FullName:       req.GetFullName(),
			Email:          req.GetEmail(),
		},
		AfterCreate: func(user sqlc.User) error {
			taskPayload := &worker.PayloadSenderVerifyEmail{
				Username: user.Username,
			}
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...) // if this return an error rollback the commit

		},
	}

	user, err := server.store.CreateUserTx(ctx, arg)
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
		User: convertUser(user.User),
	}
	return res, nil

}
func validateCreateUserRequest(req *pb.CreateUserRequest) (violation []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, ErrorResponse("username", err))
	}
	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violation = append(violation, ErrorResponse("password", err))
	}
	if err := val.ValidateFullname(req.GetFullName()); err != nil {
		violation = append(violation, ErrorResponse("full_name", err))
	}
	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violation = append(violation, ErrorResponse("email", err))
	}

	return violation
}
