package access

import (
	"context"

	accessv1 "github.com/defany/auth-service/app/pkg/gen/proto/access/v1"
	"github.com/defany/chat-server/app/internal/client"
)

type Client struct {
	proto accessv1.AccessServiceClient
}

func NewClient(proto accessv1.AccessServiceClient) client.Access {
	return &Client{
		proto: proto,
	}
}

func (c *Client) Check(ctx context.Context, endpoint string) error {
	_, err := c.proto.Check(ctx, &accessv1.CheckRequest{
		Endpoint: endpoint,
	})
	if err != nil {
		return err
	}

	return nil
}
