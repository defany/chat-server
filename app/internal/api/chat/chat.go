package chat

import (
	servicedef "github.com/defany/chat-server/app/internal/service"
	"github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"log/slog"
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
