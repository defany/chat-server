package chat

import (
	"log/slog"

	servicedef "github.com/defany/chat-server/app/internal/service"
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
)

type Implementation struct {
	chatv1.UnimplementedChatServer

	log *slog.Logger

	service servicedef.Chat
}

func NewImplementation(log *slog.Logger, service servicedef.Chat) *Implementation {
	return &Implementation{
		log:     log,
		service: service,
	}
}
