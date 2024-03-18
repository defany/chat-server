package chatrepo

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/defany/slogger/pkg/logger/sl"
)

func (r *repository) Delete(ctx context.Context, id int64) error {
	op := sl.FnName()

	q := r.qb.Delete(chats).
		Where(squirrel.Eq{
			chatsID: id,
		})

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
