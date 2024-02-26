package chatservice

import (
	"context"
	"github.com/defany/chat-server/app/internal/converter"
	"github.com/defany/chat-server/app/internal/model"
	"github.com/defany/chat-server/app/pkg/logger/sl"
)

func (s *service) SendMessage(ctx context.Context, input converter.SendMessageInput) error {
	op := sl.FnName()

	err := s.tx.ReadCommitted(ctx, func(ctx context.Context) error {
		err := s.repo.SendMessage(ctx, input)
		if err != nil {
			return err
		}

		err = s.log.Log(ctx, model.Log{
			Action: model.LogSendMessage,
			UserID: input.From,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sl.Err(op, err)
	}

	return nil
}
