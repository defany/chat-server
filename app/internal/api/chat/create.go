package chat

import (
	"context"
	"github.com/defany/chat-server/app/internal/converter"
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"github.com/defany/chat-server/app/pkg/logger/sl"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

func (i *Implementation) Create(ctx context.Context, request *chatv1.CreateRequest) (*chatv1.CreateResponse, error) {
	log := i.log.With(slog.String("op", sl.FnName()))

	output, err := i.service.CreateChat(ctx, converter.ToCreateChatInput(request))
	if err != nil {
		log.Error("failed to create chat", sl.ErrAttr(err))

		return nil, status.Error(codes.Internal, "failed to create chat")
	}

	return converter.FromCreateChatInput(output), nil
}
