package app

import (
	"context"
	"fmt"
	"log"
	"net"

	accessv1 "github.com/defany/auth-service/app/pkg/gen/proto/access/v1"
	"github.com/defany/chat-server/app/internal/interceptor"
	"github.com/defany/chat-server/app/pkg/closer"
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type App struct {
	di           *DI
	grpcServer   *grpc.Server
	accessClient accessv1.AccessServiceClient
}

func NewApp() *App {
	a := &App{}

	a.setupDI()

	return a
}

func (a *App) Run(ctx context.Context) error {
	defer func() {
		a.di.Log(ctx).Info("closing application... :(")

		closer.Close()

		a.di.Log(ctx).Info("application closed")
	}()

	a.setupDI()

	if err := a.registerAccessClient(ctx); err != nil {
		return err
	}

	a.registerUserService(ctx)

	return a.runGRPCServer(ctx)
}

func (a *App) DI() *DI {
	return a.di
}

func (a *App) setupDI() {
	a.di = newDI()
}

func (a *App) runGRPCServer(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.di.Config(ctx).Server.Port))
	if err != nil {
		return err
	}

	a.di.Log(ctx).Info("Go!")

	if err := a.grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (a *App) registerUserService(ctx context.Context) {
	validateMiddleware := interceptor.GRPCValidate(a.accessClient)

	a.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(validateMiddleware))
	reflection.Register(a.grpcServer)

	chatv1.RegisterChatServer(a.grpcServer, a.di.ChatImpl(ctx))

	return
}

func (a *App) registerAccessClient(ctx context.Context) error {
	log.Println(a.DI().Config(ctx).Server.AuthServerAddr)

	lis, err := grpc.Dial(a.DI().Config(ctx).Server.AuthServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	a.accessClient = accessv1.NewAccessServiceClient(lis)

	return nil
}
