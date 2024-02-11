package main

import (
	context "context"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
	"os"
)

const port = 50001

type server struct {
	chatv1.UnimplementedChatServer
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

func (s *server) Delete(ctx context.Context, request *chatv1.DeleteRequest) (*chatv1.DeleteResponse, error) {
	log := slog.With(
		slog.Int64("chat_id", request.GetId()),
	)

	log.Info("delete chat request")

	return &chatv1.DeleteResponse{}, nil
}

func (s *server) SendMessage(ctx context.Context, request *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	log := slog.With(
		slog.Int64("from", request.GetFrom()),
		slog.String("text", request.GetText()),
		slog.String("timestamp", request.GetTimestamp().String()),
	)

	log.Info("send message request")

	return &chatv1.SendMessageResponse{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("failed to listen: %v", err)

		os.Exit(1)
	}

	s := grpc.NewServer()

	reflection.Register(s)

	chatv1.RegisterChatServer(s, &server{})

	slog.Info("listening", slog.String("port", lis.Addr().String()))

	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve: %v", err)

		os.Exit(1)
	}
}