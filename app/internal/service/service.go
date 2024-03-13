package servicedef

import (
	"context"
	"github.com/defany/chat-server/app/internal/converter"
)

type Chat interface {
	CreateChat(ctx context.Context, input converter.CreateChatInput) (converter.CreateChatOutput, error)
	DeleteChat(ctx context.Context, input converter.DeleteChatInput) error
	SendMessage(ctx context.Context, input converter.SendMessageInput) error
}
