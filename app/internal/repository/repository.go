package repository

import (
	"context"
	"github.com/defany/chat-server/app/internal/converter"
	"github.com/defany/chat-server/app/internal/model"
)

type Chat interface {
	Create(ctx context.Context, chat model.Chat) error
	Delete(ctx context.Context, id int64) error
	SendMessage(ctx context.Context, input converter.SendMessageInput) error
}

type Log interface {
	Log(ctx context.Context, log model.Log) error
}
