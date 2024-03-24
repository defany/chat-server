package converter

import (
	chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"
)

type CreateChatInput struct {
	Title     string
	Nicknames []string
	UserID    uint64
}

type CreateChatOutput struct {
	ID uint64
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

func ToCreateChatInput(userID uint64, req *chatv1.CreateRequest) CreateChatInput {
	return CreateChatInput{
		Title:     req.GetTitle(),
		Nicknames: req.GetUsernames(),
		// Зафиксим это, когда будем делать авторизацию и будем брать из заголовка
		UserID: userID,
	}
}

func FromCreateChatInput(input CreateChatOutput) *chatv1.CreateResponse {
	return &chatv1.CreateResponse{
		Id: int64(input.ID),
	}
}

func ToDeleteChatInput(userID uint64, req *chatv1.DeleteRequest) DeleteChatInput {
	return DeleteChatInput{
		ChatID: req.GetId(),
		UserID: userID,
	}
}

func ToSendMessageInput(req *chatv1.SendMessageRequest) SendMessageInput {
	return SendMessageInput{
		ChatID: req.GetChatId(),
		From:   uint64(req.GetFrom()),
		Text:   req.GetText(),
	}
}
