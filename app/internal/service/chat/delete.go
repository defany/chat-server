package chatservice

import (
	"context"
	"github.com/defany/chat-server/app/internal/converter"
	"github.com/defany/chat-server/app/internal/model"
	"github.com/defany/chat-server/app/pkg/logger/sl"
)

func (s *service) DeleteChat(ctx context.Context, input converter.DeleteChatInput) error {
	op := sl.FnName()

	err := s.tx.ReadCommitted(ctx, func(ctx context.Context) error {
		err := s.repo.Delete(ctx, input.ChatID)
		if err != nil {
			return err
		}

		err = s.log.Log(ctx, model.Log{
			Action: model.LogDeleteChat,
			UserID: input.UserID,
		})

		return nil
	})
	if err != nil {
		return sl.Err(op, err)
	}

	return nil
}