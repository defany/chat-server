package model

const (
	LogCreateChat  = "create_chat"
	LogDeleteChat  = "delete_chat"
	LogSendMessage = "send_message"
)

type Log struct {
	Action string
	UserID uint64
}
