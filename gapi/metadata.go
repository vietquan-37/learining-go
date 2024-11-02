package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	xForwardedForHeader        = "x-forwarded-for"
	UserAgentHeader            = "user-agent"
)

type Metadata struct {
	UserAgent string
	ClientIp  string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}
	if meta, ok := metadata.FromIncomingContext(ctx); ok {

		if UserAgent := meta.Get(grpcGatewayUserAgentHeader); len(UserAgent) > 0 {
			mtdt.UserAgent = UserAgent[0]
		}
		if UserAgent := meta.Get(UserAgentHeader); len(UserAgent) > 0 {
			mtdt.UserAgent = UserAgent[0]
		}
		if ClientIp := meta.Get(xForwardedForHeader); len(ClientIp) > 0 {
			mtdt.ClientIp = ClientIp[0]
		}

	}
	if p, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIp = p.Addr.String()
	}
	return mtdt
}
