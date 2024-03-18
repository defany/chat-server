package usertests

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/defany/chat-server/app/internal/converter"
	"github.com/defany/chat-server/app/internal/model"
	"github.com/defany/chat-server/app/internal/repository"
	mockrepository "github.com/defany/chat-server/app/internal/repository/mocks"
	chatservice "github.com/defany/chat-server/app/internal/service/chat"
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"github.com/defany/db/pkg/postgres"
	mockpostgres "github.com/defany/db/pkg/postgres/mocks"
	"github.com/defany/slogger/pkg/logger/sl"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_SuccessChatCreate(t *testing.T) {
	type args struct {
		ctx             context.Context
		chatCreateInput converter.CreateChatInput
		logCreateInput  model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		chatID = gofakeit.Int64()

		userID = gofakeit.Uint64()

		title     = gofakeit.JobTitle()
		nicknames = []string{gofakeit.JobTitle(), gofakeit.JobTitle()}

		chatCreateInput = converter.CreateChatInput{
			Title:     title,
			Nicknames: nicknames,
			UserID:    userID,
		}

		chatCreateOutput = converter.CreateChatOutput{
			ID: uint64(chatID),
		}

		logCreateInput = model.Log{
			Action: model.LogCreateChat,
			UserID: userID,
		}
	)

	tests := []struct {
		name   string
		args   args
		want   *chatv1.CreateResponse
		err    error
		mocker func(tt args) mocker
	}{
		{
			name: "success",
			args: args{
				ctx:             context.Background(),
				chatCreateInput: chatCreateInput,
				logCreateInput:  logCreateInput,
			},
			want: converter.FromCreateChatInput(chatCreateOutput),
			err:  nil,
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

				chatRepo.On("Create", txCtx, model.Chat{
					Title: tt.chatCreateInput.Title,
				}).Return(uint64(chatID), nil)

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

			output, err := service.CreateChat(tt.args.ctx, tt.args.chatCreateInput)

			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, converter.FromCreateChatInput(output))
		})
	}
}

func TestService_FailChatCreateProcessTx(t *testing.T) {
	type args struct {
		ctx             context.Context
		chatCreateInput converter.CreateChatInput
		logCreateInput  model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		err = errors.New("failed to create chat")

		slErr = sl.Err("service.CreateChat", err)
	)

	tests := []struct {
		name   string
		args   args
		want   converter.CreateChatOutput
		err    error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to start tx because ReadCommitted returned an error",
			args: args{
				ctx:             context.Background(),
				chatCreateInput: converter.CreateChatInput{},
				logCreateInput:  model.Log{},
			},
			want: converter.CreateChatOutput{},
			err:  slErr,
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

			output, err := service.CreateChat(tt.args.ctx, tt.args.chatCreateInput)

			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, output)
		})
	}
}

func TestService_FailChatCreate(t *testing.T) {
	type args struct {
		ctx             context.Context
		chatCreateInput converter.CreateChatInput
		logCreateInput  model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		userID = gofakeit.Uint64()

		title     = gofakeit.JobTitle()
		nicknames = []string{gofakeit.JobTitle(), gofakeit.JobTitle()}

		chatCreateInput = converter.CreateChatInput{
			Title:     title,
			Nicknames: nicknames,
			UserID:    userID,
		}

		err = errors.New("failed to create chat")

		slErr = sl.Err("service.CreateChat", err)
	)

	tests := []struct {
		name   string
		args   args
		want   converter.CreateChatOutput
		err    error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to create chat because chat repository returned an error",
			args: args{
				ctx:             context.Background(),
				chatCreateInput: chatCreateInput,
				logCreateInput:  model.Log{},
			},
			want: converter.CreateChatOutput{},
			err:  slErr,
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

				chatRepo.On("Create", txCtx, model.Chat{
					Title: tt.chatCreateInput.Title,
				}).Return(userID, err)

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

			output, err := service.CreateChat(tt.args.ctx, tt.args.chatCreateInput)

			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, output)
		})
	}
}

func TestService_FailChatCreateLog(t *testing.T) {
	type args struct {
		ctx             context.Context
		chatCreateInput converter.CreateChatInput
		logCreateInput  model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		userID = gofakeit.Uint64()

		title     = gofakeit.JobTitle()
		nicknames = []string{gofakeit.JobTitle(), gofakeit.JobTitle()}

		chatCreateInput = converter.CreateChatInput{
			Title:     title,
			Nicknames: nicknames,
			UserID:    userID,
		}

		logCreateInput = model.Log{
			Action: model.LogCreateChat,
			UserID: userID,
		}

		err = errors.New("failed to create chat")

		slErr = sl.Err("service.CreateChat", err)
	)

	tests := []struct {
		name   string
		args   args
		want   converter.CreateChatOutput
		err    error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to create chat",
			args: args{
				ctx:             context.Background(),
				chatCreateInput: chatCreateInput,
				logCreateInput:  logCreateInput,
			},
			want: converter.CreateChatOutput{},
			err:  slErr,
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

				chatRepo.On("Create", txCtx, model.Chat{
					Title: tt.chatCreateInput.Title,
				}).Return(userID, nil)

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

			output, err := service.CreateChat(tt.args.ctx, tt.args.chatCreateInput)

			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, output)
		})
	}
}
