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
		send: make(chan []byte),
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
			var payload game.JoinGamePayload
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				log.Printf("failed to unmarshal payload %v", err)
				continue
			}
			fmt.Printf("player wants to join: %s\n", payload.PlayerName)
			player := c.game.JoinGame(payload.PlayerName)
			playerJoinedPayload := game.PlayerJoinedPayload{
				PlayerID:   player.PlayerID,
				PlayerName: player.PlayerName,
			}
			payloadData, err := json.Marshal(playerJoinedPayload)
			if err != nil {
				log.Printf("failed to marshal payload %v", err)
				continue
			}
			playerJoinedMessage := Message{
				Type:    PlayerJoined,
				Payload: payloadData,
			}

			messageData, err := json.Marshal(playerJoinedMessage)
			if err != nil {
				log.Printf("failed to marshal payload %v", err)
				continue
			}

			c.hub.broadcast <- messageData
			fmt.Printf("players: %+v", c.game.Players)

		default:
			log.Printf("unknown message type %s", msg.Type)
		}

	}
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

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
