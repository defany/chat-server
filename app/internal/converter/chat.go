package converter

import (
	"github.com/brianvoe/gofakeit/v6"
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
)

type CreateChatInput struct {
	Title  string
	UserID uint64
}

type CreateChatOutput struct {
	ID int64
}

type DeleteChatInput struct {
	ChatID int64
	UserID uint64
}

type SendMessageInput struct {
	ChatID int64
	From   uint64
	Text   string
}

func ToCreateChatInput(req *chatv1.CreateRequest) CreateChatInput {
	return CreateChatInput{
		Title: req.GetTitle(),
		// Зафиксим это, когда будем делать авторизацию и будем брать из заголовка
		UserID: uint64(gofakeit.Uint32()),
	}
}

func FromCreateChatInput(input CreateChatOutput) *chatv1.CreateResponse {
	return &chatv1.CreateResponse{
		Id: input.ID,
	}
}

func ToDeleteChatInput(req *chatv1.DeleteRequest) DeleteChatInput {
	return DeleteChatInput{
		ChatID: req.GetId(),
		// Зафиксим это, когда будем делать авторизацию и будем брать из заголовка
		UserID: uint64(gofakeit.Uint32()),
	}
}

func ToSendMessageInput(req *chatv1.SendMessageRequest) SendMessageInput {
	return SendMessageInput{
		ChatID: req.GetChatId(),
		From:   uint64(req.GetFrom()),
		Text:   req.GetText(),
	}
}
