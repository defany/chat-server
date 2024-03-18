package chattests

import (
	"context"
	"log/slog"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/defany/chat-server/app/internal/api/chat"
	"github.com/defany/chat-server/app/internal/converter"
	servicedef "github.com/defany/chat-server/app/internal/service"
	mockservicedef "github.com/defany/chat-server/app/internal/service/mocks"
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
	"github.com/defany/slogger/pkg/logger/handlers/slogpretty"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestImplementation_SuccessDelete(t *testing.T) {
	type mocker struct {
		service servicedef.Chat
	}

	type args struct {
		ctx context.Context
		req *chatv1.DeleteRequest
	}

	var (
		ctx = context.Background()

		id = gofakeit.Int64()

		userID = uint64(0)

		req = &chatv1.DeleteRequest{
			Id: id,
		}

		res = &emptypb.Empty{}
	)

	tests := []struct {
		name   string
		args   args
		want   *emptypb.Empty
		err    error
		mocker func(tt args) mocker
	}{
		{
			name: "success delete chat",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: res,
			err:  nil,
			mocker: func(tt args) mocker {
				service := mockservicedef.NewMockChat(t)

				service.On("DeleteChat", tt.ctx, converter.ToDeleteChatInput(userID, tt.req)).Return(nil)

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

			res, err := impl.Delete(ctx, tt.args.req)

			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, res)
		})
	}
}
