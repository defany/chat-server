package chatservice

import (
	"github.com/defany/chat-server/app/internal/repository"
	servicedef "github.com/defany/chat-server/app/internal/service"
	"github.com/defany/db/pkg/postgres"
)

type service struct {
	tx   postgres.TxManager
	repo repository.Chat
	log  repository.Log
}

func NewService(tx postgres.TxManager, repo repository.Chat, log repository.Log) servicedef.Chat {
	return &service{
		tx:   tx,
		repo: repo,
		log:  log,
	}
}
