package converter

import chatv1 "github.com/defany/chat-server/app/pkg/gen/chat/v1"

type CreateChatInput struct {
	Title string
}

type CreateChatOutput struct {
	ID int64
}

type DeleteChatInput struct {
	ChatID int64
}

type SendMessageInput struct {
	ChatID int64
	From   int64
	Text   string
}

func ToCreateChatInput(req *chatv1.CreateRequest) CreateChatInput {
	return CreateChatInput{
		Title: req.GetTitle(),
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
	}
}

func ToSendMessageInput(req *chatv1.SendMessageRequest) SendMessageInput {
	return SendMessageInput{
		ChatID: req.GetChatId(),
		From:   req.GetFrom(),
		Text:   req.GetText(),
	}
}
