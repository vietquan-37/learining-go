package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/vietquan-37/simplebank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	tokenTypeBearer     = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context, accessableRole []string) (paylod *token.Payload, err error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing the metadata")
	}
	value := meta.Get(authorizationHeader)
	if len(value) == 0 {
		return nil, fmt.Errorf("missing the authoriaztion header")
	}
	authHeader := value[0]
	field := strings.Fields(authHeader)
	if len(field) < 2 {
		return nil, fmt.Errorf("invalid authoriaztion header format")
	}
	authType := strings.ToLower(field[0])
	if authType != tokenTypeBearer {
		return nil, fmt.Errorf("unsupport token type")
	}
	accessToken := field[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, err
	}
	if !hasPermission(payload.Role, accessableRole) {
		return nil, fmt.Errorf("permission denied")
	}
	return payload, nil
}
func hasPermission(userRole string, accessableRole []string) bool {
	for _, role := range accessableRole {
		if userRole == role {
			return true
		}
	}
	return false
}
