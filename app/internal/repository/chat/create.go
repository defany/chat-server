package chatrepo

import (
	"context"

	"github.com/defany/chat-server/app/internal/model"
	"github.com/defany/slogger/pkg/logger/sl"
	"github.com/jackc/pgx/v5"
)

func (r *repository) Create(ctx context.Context, chat model.Chat) (uint64, error) {
	op := sl.FnName()

	q := r.qb.Insert(chats).
		Columns(chatsTitle).
		Values(chat.Title).
		Suffix("returning id")

	sql, args, err := q.ToSql()
	if err != nil {
		return 0, sl.Err(op, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return 0, sl.Err(op, err)
	}

	id, err := pgx.CollectOneRow(rows, pgx.RowTo[uint64])
	if err != nil {
		return 0, sl.Err(op, err)
	}

	return id, nil
}
