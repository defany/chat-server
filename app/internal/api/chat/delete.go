package chat

import (
	"context"
	"log/slog"

	"github.com/defany/chat-server/app/internal/converter"
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"github.com/defany/slogger/pkg/logger/sl"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) Delete(ctx context.Context, request *chatv1.DeleteRequest) (*emptypb.Empty, error) {
	log := i.log.With(slog.String("op", sl.FnName()))

	// FIXME: change it on real user id
	err := i.service.DeleteChat(ctx, converter.ToDeleteChatInput(0, request))
	if err != nil {
		log.Error("failed to delete chat", sl.ErrAttr(err))

		return nil, status.Error(codes.Internal, "failed to delete chat")
	}

	return &emptypb.Empty{}, nil
}
