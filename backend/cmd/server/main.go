package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/danielroschmann/interactive-quiz-game/backend/internal/game"
	"github.com/danielroschmann/interactive-quiz-game/backend/internal/game/websocket"
)

func main() {
	hub := websocket.NewHub()
	rounds, err := game.LoadRounds("../../internal/game/questions.json")
	if err != nil {
		log.Fatalf("failed to load rounds %v", err)
	}
	game := game.NewGame(rounds)
	go hub.Run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, game, w, r)
	})

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
