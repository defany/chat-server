package client

import (
	"context"
)

type Access interface {
	Check(ctx context.Context, endpoint string) error
}
