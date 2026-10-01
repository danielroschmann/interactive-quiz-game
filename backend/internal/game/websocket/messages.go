package websocket

import "encoding/json"

type MessageType string

const (
	JoinGame  MessageType = "join_game"
	LeaveGame MessageType = "leave_game"
	Buzz      MessageType = "buzz"
	Answer    MessageType = "answer"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
