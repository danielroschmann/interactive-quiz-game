package websocket

import (
	"log"
	"sync"

	"github.com/danielroschmann/interactive-quiz-game/backend/internal/game"
)

type Hub struct {
	mu         sync.Mutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	game       *game.Game
}

func NewHub(g *game.Game) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		game:       g,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				if client.disconnectPlayer() {
					msg, err := h.buildGameStateMessage(h.game)
					if err != nil {
						log.Printf("failed to build game state %v", err)
						continue
					}
					h.broadcastMessage(msg)
				}
			}
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	playerDisconnected := false
	for client := range h.clients {
		select {
		case client.send <- message:

		default:
			delete(h.clients, client)
			close(client.send)

			if client.disconnectPlayer() {
				playerDisconnected = true
			}
		}
	}

	if playerDisconnected {
		msg, err := h.buildGameStateMessage(h.game)
		if err != nil {
			log.Printf("failed to build game state %v", err)
			return
		}

		for client := range h.clients {
			select {
			case client.send <- msg:
			default:
			}
		}
	}
}
