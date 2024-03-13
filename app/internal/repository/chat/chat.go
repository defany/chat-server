package chatrepo

import (
	"github.com/Masterminds/squirrel"
	repo "github.com/defany/chat-server/app/internal/repository"
	"github.com/defany/db/pkg/postgres"
)

const (
	chats         = "chats"
	chatsMessages = "chats_messages"
)

const (
	chatsID    = "id"
	chatsTitle = "title"
)

const (
	chatsMessagesChatID = "chat_id"
	chatsMessagesFrom   = "from"
	chatsMessagesText   = "text"
)

type repository struct {
	db postgres.Postgres
	qb squirrel.StatementBuilderType
}

func NewRepository(db postgres.Postgres) repo.Chat {
	return &repository{
		db: db,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
