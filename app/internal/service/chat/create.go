package chatservice

import (
	"context"
	"github.com/defany/chat-server/app/internal/converter"
	"github.com/defany/chat-server/app/internal/model"
	"github.com/defany/chat-server/app/pkg/logger/sl"
)

func (s *service) CreateChat(ctx context.Context, input converter.CreateChatInput) (converter.CreateChatOutput, error) {
	op := sl.FnName()

	var output converter.CreateChatOutput

	err := s.tx.ReadCommitted(ctx, func(ctx context.Context) error {
		err := s.repo.Create(ctx, model.Chat{
			Title: input.Title,
		})
		if err != nil {
			return err
		}

		err = s.log.Log(ctx, model.Log{
			Action: model.LogCreateChat,
			UserID: input.UserID,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return converter.CreateChatOutput{}, sl.Err(op, err)
	}

	return output, nil
}
