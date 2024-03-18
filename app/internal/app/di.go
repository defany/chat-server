package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/defany/chat-server/app/internal/api/chat"
	"github.com/defany/chat-server/app/internal/config"
	"github.com/defany/chat-server/app/internal/repository"
	chatrepo "github.com/defany/chat-server/app/internal/repository/chat"
	logrepo "github.com/defany/chat-server/app/internal/repository/log"
	servicedef "github.com/defany/chat-server/app/internal/service"
	chatservice "github.com/defany/chat-server/app/internal/service/chat"
	"github.com/defany/chat-server/app/pkg/closer"
	"github.com/defany/db/pkg/postgres"
	"github.com/defany/slogger/pkg/logger/sl"
)

type DI struct {
	log *slog.Logger

	cfg *config.Config

	repositories struct {
		chat repository.Chat
		log  repository.Log
	}

	services struct {
		chat servicedef.Chat
	}

	implementations struct {
		chat *chat.Implementation
	}

	txManager postgres.TxManager
	db        postgres.Postgres
}

func newDI() *DI {
	return &DI{}
}

func (d *DI) Log(ctx context.Context) *slog.Logger {
	if d.log != nil {
		return d.log
	}

	d.log = sl.NewSlogLogger(d.Config(ctx).Logger)

	return d.log
}

func (d *DI) Config(_ context.Context) *config.Config {
	if d.cfg != nil {
		return d.cfg
	}

	d.cfg = config.MustLoad()

	return d.cfg
}

func (d *DI) Database(ctx context.Context) postgres.Postgres {
	if d.db != nil {
		return d.db
	}

	cfg := d.Config(ctx)

	dbConfig := postgres.NewConfig(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)

	dbConfig.WithRetryConnDelay(cfg.Database.ConnectAttemptsDelay)
	dbConfig.WithMaxConnAttempts(cfg.Database.ConnectAttempts)

	db, err := postgres.NewPostgres(ctx, d.Log(ctx), dbConfig)
	if err != nil {
		d.Log(ctx).Error("failed to connect to database", sl.ErrAttr(err))

		os.Exit(1)
	}

	closer.Add(func() error {
		db.Close()

		return nil
	})

	d.db = db

	return d.db
}

func (d *DI) TxManager(ctx context.Context) postgres.TxManager {
	if d.txManager != nil {
		return d.txManager
	}

	d.txManager = postgres.NewTxManager(d.Database(ctx))

	return d.txManager
}

func (d *DI) ChatRepo(ctx context.Context) repository.Chat {
	if d.repositories.chat != nil {
		return d.repositories.chat
	}

	d.repositories.chat = chatrepo.NewRepository(d.Database(ctx))

	return d.repositories.chat
}

func (d *DI) LogRepo(ctx context.Context) repository.Log {
	if d.repositories.log != nil {
		return d.repositories.log
	}

	d.repositories.log = logrepo.NewRepository(d.Database(ctx))

	return d.repositories.log
}

func (d *DI) ChatService(ctx context.Context) servicedef.Chat {
	if d.services.chat != nil {
		return d.services.chat
	}

	d.services.chat = chatservice.NewService(d.TxManager(ctx), d.ChatRepo(ctx), d.LogRepo(ctx))

	return d.services.chat
}

func (d *DI) ChatImpl(ctx context.Context) *chat.Implementation {
	if d.implementations.chat != nil {
		return d.implementations.chat
	}

	d.implementations.chat = chat.NewImplementation(d.Log(ctx), d.ChatService(ctx))

	return d.implementations.chat
}
