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
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"github.com/defany/db/pkg/postgres"
	mockpostgres "github.com/defany/db/pkg/postgres/mocks"
	"github.com/defany/slogger/pkg/logger/sl"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

func TestService_SuccessSendMessage(t *testing.T) {
	type args struct {
		ctx              context.Context
		sendMessageInput converter.SendMessageInput
		logCreateInput   model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		chatID = gofakeit.Int64()

		userID = gofakeit.Uint64()

		text      = gofakeit.JobTitle()
		timestamp = gofakeit.Date()

		req = &chatv1.SendMessageRequest{
			ChatId:    chatID,
			From:      int64(userID),
			Text:      text,
			Timestamp: timestamppb.New(timestamp),
		}

		sendMessageInput = converter.ToSendMessageInput(req)

		logCreateInput = model.Log{
			Action: model.LogSendMessage,
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
			name: "success send message in chat",
			args: args{
				ctx:              context.Background(),
				sendMessageInput: sendMessageInput,
				logCreateInput:   logCreateInput,
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

				chatRepo.On("SendMessage", txCtx, tt.sendMessageInput).Return(nil)

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

			err := service.SendMessage(tt.args.ctx, tt.args.sendMessageInput)

			require.Equal(t, tt.want, err)
		})
	}
}

func TestService_FailSendMessageProcessTx(t *testing.T) {
	type args struct {
		ctx              context.Context
		sendMessageInput converter.SendMessageInput
		logCreateInput   model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		err = errors.New("failed to send message in chat")

		slErr = sl.Err("service.SendMessage", err)
	)

	tests := []struct {
		name   string
		args   args
		want   error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to start tx because txManager.ReadCommitted returned an error",
			args: args{
				ctx:              context.Background(),
				sendMessageInput: converter.SendMessageInput{},
				logCreateInput:   model.Log{},
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

			err := service.SendMessage(tt.args.ctx, tt.args.sendMessageInput)

			require.Equal(t, tt.want, err)
		})
	}
}

func TestService_FailSendMessage(t *testing.T) {
	type args struct {
		ctx              context.Context
		sendMessageInput converter.SendMessageInput
		logCreateInput   model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		userID = gofakeit.Uint64()

		chatID = gofakeit.Int64()

		text      = gofakeit.JobTitle()
		timestamp = gofakeit.Date()

		req = &chatv1.SendMessageRequest{
			ChatId:    chatID,
			From:      int64(userID),
			Text:      text,
			Timestamp: timestamppb.New(timestamp),
		}

		sendMessageInput = converter.ToSendMessageInput(req)

		err = errors.New("failed to send message in chat")

		slErr = sl.Err("service.SendMessage", err)
	)

	tests := []struct {
		name   string
		args   args
		want   error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to send message in chat because chat repository returned an error",
			args: args{
				ctx:              context.Background(),
				sendMessageInput: sendMessageInput,
				logCreateInput:   model.Log{},
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

				chatRepo.On("SendMessage", txCtx, tt.sendMessageInput).Return(err)

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

			err := service.SendMessage(tt.args.ctx, tt.args.sendMessageInput)

			require.Equal(t, tt.want, err)
		})
	}
}

func TestService_FailSendMessageLog(t *testing.T) {
	type args struct {
		ctx              context.Context
		sendMessageInput converter.SendMessageInput
		logCreateInput   model.Log
	}

	type mocker struct {
		txManager postgres.TxManager
		chat      repository.Chat
		log       repository.Log
	}

	var (
		userID = gofakeit.Uint64()

		chatID = gofakeit.Int64()

		text      = gofakeit.JobTitle()
		timestamp = gofakeit.Date()

		req = &chatv1.SendMessageRequest{
			ChatId:    chatID,
			From:      int64(userID),
			Text:      text,
			Timestamp: timestamppb.New(timestamp),
		}

		sendMessageInput = converter.ToSendMessageInput(req)

		logCreateInput = model.Log{
			Action: model.LogSendMessage,
			UserID: userID,
		}

		err = errors.New("failed to send message in chat")

		slErr = sl.Err("service.SendMessage", err)
	)

	tests := []struct {
		name   string
		args   args
		want   error
		mocker func(tt args) mocker
	}{
		{
			name: "failed to send message in because log repository returned an error",
			args: args{
				ctx:              context.Background(),
				sendMessageInput: sendMessageInput,
				logCreateInput:   logCreateInput,
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

				chatRepo.On("SendMessage", txCtx, tt.sendMessageInput).Return(nil)

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

			err := service.SendMessage(tt.args.ctx, tt.args.sendMessageInput)

			require.Equal(t, tt.want, err)
		})
	}
}
