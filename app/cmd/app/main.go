package main

import (
	"context"
	"github.com/defany/chat-server/app/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := app.NewApp()

	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}
