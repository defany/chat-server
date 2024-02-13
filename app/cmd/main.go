package main

import (
	context "context"
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
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const port = 50001

const dsn = "postgres://defany:137278DfN@postgres:5432/messenger"

const (
	chats      = "chats"
	messages   = "chats_messages"
	usersChats = "users_chats"
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
		Columns("title").
		Values(gofakeit.BookTitle()).
		Suffix("returning id")

	sql, args, err := q.ToSql()
	if err != nil {
		log.Error("failed to build query to create chat", slog.String("error", err.Error()))

		return &chatv1.CreateResponse{}, ErrCreateChat
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		log.Error("failed to execute query to create chat", slog.String("error", err.Error()))

		return &chatv1.CreateResponse{}, ErrCreateChat
	}
	defer rows.Close()

	chatID, err := pgx.CollectOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		log.Error("error getting chat id", slog.String("error", err.Error()))

		return &chatv1.CreateResponse{}, ErrCreateChat
	}

	log = log.With(slog.Int64("chat_id", chatID))

	q = s.qb.Insert(usersChats).
		Columns("chat_id", "user_id")

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

		return &chatv1.CreateResponse{}, ErrCreateChat
	}

	_, err = s.db.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to exec query to add users in chat", slog.String("error", err.Error()))

		return &chatv1.CreateResponse{}, ErrCreateChat
	}

	return &chatv1.CreateResponse{
		Id: chatID,
	}, nil
}

/*
	Возможно, лучше было бы добавить поле is_deleted или вообще status для чата и не удалять его,
	а просто менять значение и не отображать в случае чего, но это, наверное, выходит за рамки и пошел просто за удаление сообщений
*/

func (s *server) Delete(ctx context.Context, request *chatv1.DeleteRequest) (*emptypb.Empty, error) {
	log := slog.With(
		slog.Int64("chat_id", request.GetId()),
	)

	log.Info("delete chat request")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction for delete chat", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrDeleteChat
	}

	q := s.qb.Delete(usersChats).
		Where(squirrel.Eq{"chat_id": request.GetId()})

	sql, args, err := q.ToSql()
	if err != nil {
		log.Error("failed to build query to delete users from chat", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrDeleteChat
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to delete users from chat", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrDeleteChat
	}

	q = s.qb.Delete(messages).
		Where(squirrel.Eq{"chat_id": request.GetId()})

	sql, args, err = q.ToSql()
	if err != nil {
		log.Error("failed to build query to delete chat messages", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrDeleteChat
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to delete chat messages", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrDeleteChat
	}

	q = s.qb.Delete(chats).
		Where(squirrel.Eq{"id": request.GetId()})

	sql, args, err = q.ToSql()
	if err != nil {
		log.Error("failed to build query to delete chat", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrDeleteChat
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to delete chat", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrDeleteChat
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction for delete chat", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrDeleteChat
	}

	return &emptypb.Empty{}, nil
}

func (s *server) SendMessage(ctx context.Context, request *chatv1.SendMessageRequest) (*emptypb.Empty, error) {
	log := slog.With(
		slog.Int64("from", request.GetFrom()),
		slog.String("text", request.GetText()),
		slog.String("timestamp", request.GetTimestamp().AsTime().String()),
	)

	log.Info("send message request")

	q := s.qb.Insert(messages).
		Columns("chat_id", "user_id", "text", "timestamp").
		Values(request.GetChatId(), request.GetFrom(), request.GetText(), request.GetTimestamp().AsTime())

	sql, args, err := q.ToSql()
	if err != nil {
		log.Error("failed to build query to send message to chat", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrCreateMessage
	}

	log.Debug(sql)

	_, err = s.db.Exec(ctx, sql, args...)
	if err != nil {
		log.Error("failed to send message to chat", slog.String("error", err.Error()))

		return &emptypb.Empty{}, ErrCreateMessage
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
