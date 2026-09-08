package websocket

type MessageType string

const (
	JoinGame  MessageType = "join_game"
	LeaveGame MessageType = "leave_game"
	Buzz      MessageType = "buzz"
	Answer    MessageType = "answer"
)
