package interceptor

import (
	"context"
	"errors"

	"github.com/defany/chat-server/app/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Interceptor struct {
	accessClient client.Access
}

func NewInterceptor(accessClient client.Access) *Interceptor {
	return &Interceptor{
		accessClient: accessClient,
	}
}

func (i *Interceptor) Interceptor(ctx context.Context, req any, server *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	outCtx, err := i.injectAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = i.accessClient.Check(outCtx, server.FullMethod)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (i *Interceptor) injectAuthHeader(ctx context.Context) (context.Context, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	authHeaderData := md.Get("Authorization")
	if len(authHeaderData) == 0 {
		return nil, errors.New("auth header is not provided")
	}

	outCtx := metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"Authorization": authHeaderData[0],
	}))

	return outCtx, nil
}
