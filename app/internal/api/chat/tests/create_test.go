package chattests

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/defany/chat-server/app/internal/api/chat"
	"github.com/defany/chat-server/app/internal/converter"
	servicedef "github.com/defany/chat-server/app/internal/service"
	mockservicedef "github.com/defany/chat-server/app/internal/service/mocks"
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"github.com/defany/slogger/pkg/logger/handlers/slogpretty"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

func TestImplementation_SuccessCreate(t *testing.T) {
	type mocker struct {
		service servicedef.Chat
	}

	type args struct {
		ctx context.Context
		req *chatv1.CreateRequest
	}

	var (
		ctx = context.Background()

		id = gofakeit.Int64()

		userID = uint64(0)

		title     = gofakeit.BookTitle()
		usernames = []string{gofakeit.BookTitle(), gofakeit.BookTitle(), gofakeit.BookTitle()}

		req = &chatv1.CreateRequest{
			Title:     title,
			Usernames: usernames,
		}

		output = converter.CreateChatOutput{
			ID: uint64(id),
		}

		res = &chatv1.CreateResponse{
			Id: id,
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
			name: "success create chat",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: res,
			err:  nil,
			mocker: func(tt args) mocker {
				service := mockservicedef.NewMockChat(t)

				service.On("CreateChat", tt.ctx, converter.ToCreateChatInput(userID, tt.req)).Return(output, nil)

				return mocker{
					service: service,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocker := tt.mocker(tt.args)

			impl := chat.NewImplementation(slog.New(slogpretty.NewHandler()), mocker.service)

			res, err := impl.Create(ctx, tt.args.req)

			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, res)
		})
	}
}
