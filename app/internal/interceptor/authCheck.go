package interceptor

import (
	"context"
	"errors"

	accessClient "github.com/defany/auth-service/app/pkg/gen/proto/access/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func GRPCValidate(client accessClient.AccessServiceClient) func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, server *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		md, _ := metadata.FromIncomingContext(ctx)

		authHeaderData := md.Get("Authorization")
		if len(authHeaderData) == 0 {
			return nil, errors.New("auth header is not provided")
		}

		outCtx := metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"Authorization": authHeaderData[0],
		}))

		_, err = client.Check(outCtx, &accessClient.CheckRequest{
			Endpoint: server.FullMethod,
		})
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}
