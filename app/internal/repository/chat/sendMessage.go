package chatrepo

import (
	"context"

	"github.com/defany/chat-server/app/internal/converter"
	"github.com/defany/slogger/pkg/logger/sl"
)

func (r *repository) SendMessage(ctx context.Context, input converter.SendMessageInput) error {
	op := sl.FnName()

	q := r.qb.Insert(chatsMessages).
		Columns(chatsMessagesChatID, chatsMessagesFrom, chatsMessagesText).
		Values(input.ChatID, input.From, input.Text)

	sql, args, err := q.ToSql()
	if err != nil {
		return sl.Err(op, err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return sl.Err(op, err)
	}

	return nil
}
