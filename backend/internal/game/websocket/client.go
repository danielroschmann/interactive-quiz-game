// WebSocket package handles connections from new clients and upgrades connection from HTTP to WS connection.
package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/danielroschmann/interactive-quiz-game/backend/internal/game"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:5173"
	},
}

type Client struct {
	hub      *Hub
	game     *game.Game
	conn     *websocket.Conn
	send     chan []byte
	playerID uint
}

func NewClient(hub *Hub, game *game.Game, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		game: game,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

func ServeWs(hub *Hub, game *game.Game, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to setup websocket", err)
		return
	}

	client := NewClient(hub, game, conn)
	hub.register <- client
	log.Println("Client connected", client)

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.leaveGame()
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Printf("failed to unmarshal message %v", err)
			continue
		}
		fmt.Printf("Type: %s", msg.Type)

		switch msg.Type {
		case JoinGame:
			c.joinGame(msg)
		case LeaveGame:
			c.leaveGame()
		case Buzz:
			c.handleBuzz()
		case Answer:
			c.handleAnswer(msg)
		case StartGame:
			c.handleStartGame()
		default:
			log.Printf("unknown message type %s", msg.Type)
		}

	}
}

func (c *Client) handleStartGame() {
	if c.playerID == 0 {
		log.Print("client has not joined the game")
		return
	}

	c.hub.mu.Lock()
	if c.playerID != c.game.HostPlayerID {
		log.Printf("client player id %d is not the host", c.playerID)
		c.hub.mu.Unlock()
		return
	}

	err := c.game.StartGame()
	c.hub.mu.Unlock()
	if err != nil {
		log.Printf("could not start game %v", err)
		return
	}

	if err := c.hub.BroadcastGameState(c.game); err != nil {
		log.Printf("could not broadcast game state %v", err)
	}
}

func (c *Client) joinGame(msg Message) {
	var payload game.JoinGamePayload
	err := json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		log.Printf("failed to unmarshal payload %v", err)
		return
	}
	fmt.Printf("player wants to join: %s\n", payload.PlayerName)
	c.hub.mu.Lock()
	player := c.game.JoinGame(payload.PlayerName)
	if c.game.HostPlayerID == 0 {
		c.game.HostPlayerID = player.PlayerID
	}
	c.hub.mu.Unlock()
	c.playerID = player.PlayerID
	playerJoinedPayload := game.PlayerJoinedPayload{
		PlayerID:   player.PlayerID,
		PlayerName: player.PlayerName,
	}
	payloadData, err := json.Marshal(playerJoinedPayload)
	if err != nil {
		log.Printf("failed to marshal payload %v", err)
		return
	}
	playerJoinedMessage := Message{
		Type:    PlayerJoined,
		Payload: payloadData,
	}

	messageData, err := json.Marshal(playerJoinedMessage)
	if err != nil {
		log.Printf("failed to marshal payload %v", err)
		return
	}

	c.hub.broadcast <- messageData
	if err := c.hub.BroadcastGameState(c.game); err != nil {
		log.Printf("failed to broadcast game state %v", err)
	}
}

func (c *Client) handleAnswer(msg Message) {
	if c.playerID == 0 {
		return
	}

	var answerPayload AnswerPayload
	if err := json.Unmarshal(msg.Payload, &answerPayload); err != nil {
		log.Printf("failed to unmarshal payload %v", err)
		return
	}

	c.hub.mu.Lock()
	correct, err := c.game.SubmitAnswer(c.playerID, answerPayload.Answer)
	c.hub.mu.Unlock()
	if err != nil {
		log.Printf("failed to submit answer %v", err)
		return
	}

	log.Printf("answer was %t", correct)
	if err := c.hub.BroadcastGameState(c.game); err != nil {
		log.Printf("failed to broadcast game state %v", err)
	}
}

func (c *Client) handleBuzz() {
	if c.playerID == 0 {
		return
	}

	c.hub.mu.Lock()
	err := c.game.Buzz(c.playerID)
	c.hub.mu.Unlock()
	if err != nil {
		log.Printf("player %d failed to buzz: %v", c.playerID, err)
	}

	if err := c.hub.BroadcastGameState(c.game); err != nil {
		log.Printf("failed to broadcast game state %v", err)
		return
	}
}

func (c *Client) leaveGame() {
	if c.playerID == 0 {
		return
	}
	c.hub.mu.Lock()
	player, ok := c.game.LeaveGame(c.playerID)
	c.hub.mu.Unlock()
	if !ok {
		return
	}
	c.playerID = 0
	playerLeftPayload := game.PlayerLeftPayload{
		PlayerID:   player.PlayerID,
		PlayerName: player.PlayerName,
	}
	payloadData, err := json.Marshal(playerLeftPayload)
	if err != nil {
		log.Printf("failed to marshal payload %v", err)
		return
	}

	playerLeftMessage := Message{
		Type:    PlayerLeft,
		Payload: payloadData,
	}

	messageData, err := json.Marshal(playerLeftMessage)
	if err != nil {
		log.Printf("failed to marshal payload %v", err)
		return
	}
	c.hub.broadcast <- messageData
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}
