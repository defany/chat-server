package chatrepo

import (
	"context"
	"github.com/defany/chat-server/app/internal/model"
	"github.com/defany/chat-server/app/pkg/logger/sl"
)

func (r *repository) Create(ctx context.Context, chat model.Chat) error {
	op := sl.FnName()

	q := r.qb.Insert(chats).
		Columns(chatsTitle).
		Values(chat.Title)

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
