package gapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/pb"
	"github.com/vietquan-37/simplebank/util"
	"github.com/vietquan-37/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unAuthorizeError(err)
	}
	violation := validateUpdateUserRequest(req)
	if violation != nil {
		return nil, invalidArgumentError(violation)
	}
	if authPayload.Username != req.GetUsername() {
		return nil, status.Error(codes.PermissionDenied, "cannot update this user")
	}
	arg := sqlc.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}
	if req.Password != nil {
		hashPassword, err := util.HashedPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "fail to hashed password: %s", err)
		}
		arg.HashedPassword = pgtype.Text{
			String: hashPassword,
			Valid:  true,
		}
		arg.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}

	}
	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, sqlc.ErrRecordNoFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "fail to register user: %s", err)
	}
	res := &pb.UpdateUserResponse{
		User: convertUser(user),
	}
	return res, nil

}
func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violation []*errdetails.BadRequest_FieldViolation) {

	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, ErrorResponse("username", err))
	}
	if req.Password != nil {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violation = append(violation, ErrorResponse("password", err))
		}
	}
	if req.FullName != nil {

		if err := val.ValidateFullname(req.GetFullName()); err != nil {
			violation = append(violation, ErrorResponse("full_name", err))
		}
	}
	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violation = append(violation, ErrorResponse("email", err))
		}
	}
	return violation
}
