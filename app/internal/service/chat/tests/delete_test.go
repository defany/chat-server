package usertests

import (
	"context"
	"errors"
	"github.com/brianvoe/gofakeit"
	"github.com/defany/chat-server/app/internal/converter"
	"github.com/defany/chat-server/app/internal/model"
	"github.com/defany/chat-server/app/internal/repository"
	mockrepository "github.com/defany/chat-server/app/internal/repository/mocks"
	chatservice "github.com/defany/chat-server/app/internal/service/chat"
	"github.com/defany/db/pkg/postgres"
	mockpostgres "github.com/defany/db/pkg/postgres/mocks"
	"github.com/defany/slogger/pkg/logger/sl"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_SuccessChatDelete(t *testing.T) {
	type args struct {
		ctx             context.Context
		deleteChatInput converter.DeleteChatInput
		logCreateInput  model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		userID = gofakeit.Uint64()
		chatID = gofakeit.Int64()

		deleteChatInput = converter.DeleteChatInput{
			ChatID: chatID,
			UserID: userID,
		}

		logCreateInput = model.Log{
			Action: model.LogDeleteChat,
			UserID: userID,
		}
	)

	tests := []struct {
		name   string
		args   args
		want   error
		mocker func(tt args) mocker
	}{
		{
			name: "success delete chat by id",
			args: args{
				ctx:             context.Background(),
				deleteChatInput: deleteChatInput,
				logCreateInput:  logCreateInput,
			},
			want: nil,
			mocker: func(tt args) mocker {
				txOpts := pgx.TxOptions{
					IsoLevel: pgx.ReadCommitted,
				}

				tx := mockpostgres.NewMockTx(t)

				txCtx := postgres.InjectTX(tt.ctx, tx)

				tx.On("Commit", txCtx).Return(nil)

				db := mockpostgres.NewMockPostgres(t)
				db.On("BeginTx", tt.ctx, txOpts).Return(tx, nil)

				txManager := postgres.NewTxManager(db)
				chatRepo := mockrepository.NewMockChat(t)
				logRepo := mockrepository.NewMockLog(t)

				chatRepo.On("Delete", txCtx, tt.deleteChatInput.ChatID).Return(nil)

				logRepo.On("Log", txCtx, tt.logCreateInput).Return(nil)

				return mocker{
					txManager: txManager,
					chat:      chatRepo,
					log:       logRepo,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Parallel()

		t.Run(tt.name, func(t *testing.T) {
			mocker := tt.mocker(tt.args)

			service := chatservice.NewService(mocker.txManager, mocker.chat, mocker.log)

			err := service.DeleteChat(tt.args.ctx, tt.args.deleteChatInput)

			require.Equal(t, tt.want, err)
		})
	}
}

func TestService_FailChatDeleteProcessTx(t *testing.T) {
	type args struct {
		ctx             context.Context
		deleteChatInput converter.DeleteChatInput
		logCreateInput  model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		err = errors.New("failed to delete chat")

		slErr = sl.Err("service.DeleteChat", err)
	)

	tests := []struct {
		name   string
		args   args
		want   error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to start tx because ReadCommitted returned an error",
			args: args{
				ctx:             context.Background(),
				deleteChatInput: converter.DeleteChatInput{},
				logCreateInput:  model.Log{},
			},
			want: slErr,
			mocker: func(tt args) mocker {
				txManager := mockpostgres.NewMockTxManager(t)
				txManager.On("ReadCommitted", tt.ctx, mock.AnythingOfType("postgres.Handler")).Return(err)

				return mocker{
					txManager: txManager,
					chat:      nil,
					log:       nil,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Parallel()

		t.Run(tt.name, func(t *testing.T) {
			mocker := tt.mocker(tt.args)

			service := chatservice.NewService(mocker.txManager, mocker.chat, mocker.log)

			err := service.DeleteChat(tt.args.ctx, tt.args.deleteChatInput)

			require.Equal(t, tt.want, err)
		})
	}
}

func TestService_FailChatDelete(t *testing.T) {
	type args struct {
		ctx             context.Context
		chatDeleteInput converter.DeleteChatInput
		logCreateInput  model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		userID = gofakeit.Uint64()
		chatID = gofakeit.Int64()

		chatDeleteInput = converter.DeleteChatInput{
			ChatID: chatID,
			UserID: userID,
		}

		err = errors.New("failed to delete chat")

		slErr = sl.Err("service.DeleteChat", err)
	)

	tests := []struct {
		name   string
		args   args
		want   error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to delete chat because chat repository returned an error",
			args: args{
				ctx:             context.Background(),
				chatDeleteInput: chatDeleteInput,
				logCreateInput:  model.Log{},
			},
			want: slErr,
			mocker: func(tt args) mocker {
				txOpts := pgx.TxOptions{
					IsoLevel: pgx.ReadCommitted,
				}

				tx := mockpostgres.NewMockTx(t)

				txCtx := postgres.InjectTX(tt.ctx, tx)

				tx.On("Rollback", txCtx).Return(nil)

				db := mockpostgres.NewMockPostgres(t)
				db.On("BeginTx", tt.ctx, txOpts).Return(tx, nil)

				txManager := postgres.NewTxManager(db)
				chatRepo := mockrepository.NewMockChat(t)
				logRepo := mockrepository.NewMockLog(t)

				chatRepo.On("Delete", txCtx, tt.chatDeleteInput.ChatID).Return(err)

				return mocker{
					txManager: txManager,
					chat:      chatRepo,
					log:       logRepo,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Parallel()

		t.Run(tt.name, func(t *testing.T) {
			mocker := tt.mocker(tt.args)

			service := chatservice.NewService(mocker.txManager, mocker.chat, mocker.log)

			err := service.DeleteChat(tt.args.ctx, tt.args.chatDeleteInput)

			require.Equal(t, tt.want, err)
		})
	}
}

func TestService_FailChatDeleteLog(t *testing.T) {
	type args struct {
		ctx             context.Context
		deleteChatInput converter.DeleteChatInput
		logCreateInput  model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		userID = gofakeit.Uint64()
		chatID = gofakeit.Int64()

		deleteChatInput = converter.DeleteChatInput{
			ChatID: chatID,
			UserID: userID,
		}

		logCreateInput = model.Log{
			Action: model.LogDeleteChat,
			UserID: userID,
		}

		err = errors.New("failed to delete chat")

		slErr = sl.Err("service.DeleteChat", err)
	)

	tests := []struct {
		name   string
		args   args
		want   error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to delete chat by id",
			args: args{
				ctx:             context.Background(),
				deleteChatInput: deleteChatInput,
				logCreateInput:  logCreateInput,
			},
			want: slErr,
			mocker: func(tt args) mocker {
				txOpts := pgx.TxOptions{
					IsoLevel: pgx.ReadCommitted,
				}

				tx := mockpostgres.NewMockTx(t)

				txCtx := postgres.InjectTX(tt.ctx, tx)

				tx.On("Rollback", txCtx).Return(nil)

				db := mockpostgres.NewMockPostgres(t)
				db.On("BeginTx", tt.ctx, txOpts).Return(tx, nil)

				txManager := postgres.NewTxManager(db)
				chatRepo := mockrepository.NewMockChat(t)
				logRepo := mockrepository.NewMockLog(t)

				chatRepo.On("Delete", txCtx, tt.deleteChatInput.ChatID).Return(nil)

				logRepo.On("Log", txCtx, tt.logCreateInput).Return(err)

				return mocker{
					txManager: txManager,
					chat:      chatRepo,
					log:       logRepo,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Parallel()

		t.Run(tt.name, func(t *testing.T) {
			mocker := tt.mocker(tt.args)

			service := chatservice.NewService(mocker.txManager, mocker.chat, mocker.log)

			err := service.DeleteChat(tt.args.ctx, tt.args.deleteChatInput)

			require.Equal(t, tt.want, err)
		})
	}
}
