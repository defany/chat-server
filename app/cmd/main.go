package main

import (
	context "context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const port = 50001

const dsn = "postgres://defany:137278DfN@postgres:5432/messenger"

type server struct {
	chatv1.UnimplementedChatServer

	db *pgxpool.Pool
}

func (s *server) Create(ctx context.Context, request *chatv1.CreateRequest) (*chatv1.CreateResponse, error) {
	chatID := gofakeit.Int64()

	log := slog.With(
		slog.Any("usernames", request.GetUsernames()),
		slog.Int64("chat_id", chatID),
	)

	log.Info("create chat request")

	return &chatv1.CreateResponse{
		Id: chatID,
	}, nil
}

func (s *server) Delete(ctx context.Context, request *chatv1.DeleteRequest) (*emptypb.Empty, error) {
	log := slog.With(
		slog.Int64("chat_id", request.GetId()),
	)

	log.Info("delete chat request")

	return &emptypb.Empty{}, nil
}

func (s *server) SendMessage(ctx context.Context, request *chatv1.SendMessageRequest) (*emptypb.Empty, error) {
	log := slog.With(
		slog.Int64("from", request.GetFrom()),
		slog.String("text", request.GetText()),
		slog.String("timestamp", request.GetTimestamp().String()),
	)

	log.Info("send message request")

	return &emptypb.Empty{}, nil
}

func main() {
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
	})

	slog.Info("listening", slog.String("port", lis.Addr().String()))

	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve: %v", err)
		os.Exit(1)
	}
}
