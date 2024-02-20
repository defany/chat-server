package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const port = 50001

const dsn = "postgres://defany:137278DfN@postgres:5432/messenger"

const (
	chats      = "chats"
	messages   = "chats_messages"
	usersChats = "users_chats"
)

const (
	UsersChatsChatID = "chat_id"
)

const (
	MessagesID        = "id"
	MessagesChatID    = "chat_id"
	MessagesUserID    = "user_id"
	MessagesText      = "text"
	MessagesTimestamp = "timestamp"
)

const (
	ChatsID    = "id"
	ChatsTitle = "title"
)

var (
	ErrCreateChat    = errors.New("failed to create chat")
	ErrDeleteChat    = errors.New("failed to delete chat")
	ErrCreateMessage = errors.New("failed to create chat message")
)

type server struct {
	chatv1.UnimplementedChatServer

	db *pgxpool.Pool
	qb squirrel.StatementBuilderType
}

func (s *server) Create(ctx context.Context, request *chatv1.CreateRequest) (*chatv1.CreateResponse, error) {
	log := slog.With(
		slog.Any("usernames", request.GetUsernames()),
	)

	log.Info("create chat request")

	q := s.qb.Insert(chats).
		Columns(ChatsTitle).
		Values(request.GetTitle()).
		Suffix("returning id")

	sql, args, err := q.ToSql()
	if err != nil {
		log.Error("failed to build query to create chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrCreateChat.Error())
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		log.Error("failed to execute query to create chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrCreateChat.Error())
	}
	defer rows.Close()

	chatID, err := pgx.CollectOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		log.Error("error getting chat id", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrCreateChat.Error())
	}

	log = log.With(slog.Int64(MessagesChatID, chatID))

	q = s.qb.Insert(usersChats).
		Columns(MessagesChatID, MessagesUserID)

	/*
		В будущем для auth сервиса прикрутим ручку, чтобы по имени юзеров получать их айди
		Пока что представляем, что мы их получили
	*/
	for _, nick := range request.GetUsernames() {
		id := int64(len(nick)) + int64(gofakeit.Uint8())

		log.Debug("adding user to chat", slog.Int64("user_id", id))

		q = q.Values(chatID, id)
	}

	sql, args, err = q.ToSql()
	if err != nil {
		log.Error("failed to build query to add users in chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrCreateChat.Error())
	}

	_, err = s.db.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to exec query to add users in chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrCreateChat.Error())
	}

	return &chatv1.CreateResponse{
		Id: chatID,
	}, nil
}

/*
	Возможно, лучше было бы добавить поле is_deleted или вообще status для чата и не удалять его,
	а просто менять значение и не отображать в случае чего, но это, наверное, выходит за рамки и пошел просто за удаление сообщений
*/

func (s *server) Delete(ctx context.Context, request *chatv1.DeleteRequest) (empty *emptypb.Empty, err error) {
	log := slog.With(
		slog.Int64("chat_id", request.GetId()),
	)

	log.Info("delete chat request")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction for delete chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrDeleteChat.Error())
	}

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(ctx); txErr != nil {
				log.Error("failed to rollback transaction for delete chat", slog.String("error", txErr.Error()))

				err = status.Error(codes.Internal, ErrDeleteChat.Error())
			}

			return
		}

		if txErr := tx.Commit(ctx); txErr != nil {
			log.Error("failed to commit transaction for delete chat", slog.String("error", txErr.Error()))

			err = status.Error(codes.Internal, ErrDeleteChat.Error())
		}

		return
	}()

	q := s.qb.Delete(usersChats).
		Where(squirrel.Eq{
			UsersChatsChatID: request.GetId(),
		})

	sql, args, err := q.ToSql()
	if err != nil {
		log.Error("failed to build query to delete users from chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrDeleteChat.Error())
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to delete users from chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrDeleteChat.Error())
	}

	q = s.qb.Delete(messages).
		Where(squirrel.Eq{
			MessagesChatID: request.GetId(),
		})

	sql, args, err = q.ToSql()
	if err != nil {
		log.Error("failed to build query to delete chat messages", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrDeleteChat.Error())
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to delete chat messages", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrDeleteChat.Error())
	}

	q = s.qb.Delete(chats).
		Where(squirrel.Eq{
			ChatsID: request.GetId(),
		})

	sql, args, err = q.ToSql()
	if err != nil {
		log.Error("failed to build query to delete chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrDeleteChat.Error())
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to delete chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrDeleteChat.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *server) SendMessage(ctx context.Context, request *chatv1.SendMessageRequest) (*emptypb.Empty, error) {
	log := slog.With(
		slog.Int64(MessagesUserID, request.GetFrom()),
		slog.String(MessagesText, request.GetText()),
		slog.String(MessagesTimestamp, request.GetTimestamp().AsTime().String()),
	)

	log.Info("send message request")

	q := s.qb.Insert(messages).
		Columns(MessagesChatID, MessagesUserID, MessagesText, MessagesTimestamp).
		Values(request.GetChatId(), request.GetFrom(), request.GetText(), request.GetTimestamp().AsTime())

	sql, args, err := q.ToSql()
	if err != nil {
		log.Error("failed to build query to send message to chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrCreateMessage.Error())
	}

	_, err = s.db.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to send message to chat", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, ErrCreateMessage.Error())
	}

	return &emptypb.Empty{}, nil
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("failed to listen: %v", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	chatv1.RegisterChatServer(s, &server{
		db: pool,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	})

	slog.Info("listening", slog.String("port", lis.Addr().String()))

	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve: %v", err)
		os.Exit(1)
	}
}
