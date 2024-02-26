package chatservice

import (
	"context"
	"github.com/defany/chat-server/app/internal/converter"
)

func (s *service) SendMessage(ctx context.Context, input converter.SendMessageInput) error {
	err := s.repo.SendMessage(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
