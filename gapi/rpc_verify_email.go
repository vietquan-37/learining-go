package gapi

import (
	"context"

	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/pb"
	"github.com/vietquan-37/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {

	violation := validateVerifyEmail(req)
	if violation != nil {
		return nil, invalidArgumentError(violation)
	}

	verify, err := server.store.VerifyEmailTx(ctx, sqlc.VerifyEmailTxParams{
		ID:         req.GetId(),
		SecretCode: req.GetSecretCode(),
	})
	if err != nil {

		return nil, status.Errorf(codes.Internal, "fail verify email: %s", err)
	}

	res := &pb.VerifyEmailResponse{
		IsVerified: verify.User.IsEmailVerified,
	}

	return res, nil

}
func validateVerifyEmail(req *pb.VerifyEmailRequest) (violation []*errdetails.BadRequest_FieldViolation) {

	if err := val.ValidateEmailId(req.GetId()); err != nil {
		violation = append(violation, ErrorResponse("id", err))
	}

	if err := val.ValidateSecretCode(req.GetSecretCode()); err != nil {
		violation = append(violation, ErrorResponse("secret_code", err))
	}

	return violation
}
