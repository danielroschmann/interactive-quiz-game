package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/danielroschmann/interactive-quiz-game/backend/internal/game"
	gorilla "github.com/gorilla/websocket"
)

func initializeGame(t *testing.T) *game.Game {
	t.Helper()

	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}

	g := game.NewGame(rounds)

	g.JoinGame("John")
	g.JoinGame("Doe")
	g.JoinGame("Foo")
	g.JoinGame("Bar")
	g.JoinGame("Troy")

	if err := g.StartGame(); err != nil {
		t.Fatalf("expected game to start, got %v", err)
	}

	return g
}

func readGameInProgress(t *testing.T, conn *gorilla.Conn) GameStatePayload {
	t.Helper()

	for {
		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatalf("could not set read deadline: %v", err)
		}

		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("could not read message %v", err)
		}

		if msg.Type != GameStateUpdate {
			continue
		}

		var payload GameStatePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			t.Fatalf("could not unmarshal game state payload: %v", err)
		}

		if payload.State == string(game.GameStateInProgress) {
			return payload
		}
	}
}

func TestBuildGameStatePayload_MultipleChoiceRound(t *testing.T) {
	game := initializeGame(t)

	game.CurrentRound = 1
	game.CurrentItem = 0

	payload, err := BuildGameStatePayload(game)
	if err != nil {
		t.Fatalf("could not build game state payload %v", err)
	}

	if payload.State != "in_progress" {
		t.Errorf("expected state in_progress, got %s", payload.State)
	}

	if payload.CurrentRound != 1 {
		t.Errorf("expected current round 1, got %d", payload.CurrentRound)
	}

	if payload.CurrentItem != 0 {
		t.Errorf("expected current item 0, got %d", payload.CurrentItem)
	}

	if payload.RoundType != "multiple_choice" {
		t.Errorf("expected round type multiple_choice, got %s", payload.RoundType)
	}

	playersPayload := len(payload.Players)
	teamsPayload := len(payload.Teams)

	if playersPayload != len(game.Players) {
		t.Errorf("expected same amount of players, got %d and %d", playersPayload, len(game.Players))
	}

	if teamsPayload != len(game.Teams) {
		t.Errorf("expected same amount of teams, got %d and %d", teamsPayload, len(game.Teams))
	}

	if payload.Question == nil {
		t.Fatalf("expected a question in payload")
	}

	if payload.Challenge != nil {
		t.Errorf("expected challenge to be nil")
	}

	question, err := game.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get question %v", err)
	}

	if payload.Question.Question != question.Question {
		t.Errorf("expected question to be the same, got %s and %s", payload.Question.Question, question.Question)
	}

	if !slices.Equal(payload.Question.AnswerOptions, question.AnswerOptions) {
		t.Errorf("expected answer options to be the same")
	}

	p, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}

	if bytes.Contains(p, []byte(`"correct_answer"`)) {
		t.Errorf("payload must not expose the correct answer")
	}
}

func TestBuildGameStatePayload_Lobby(t *testing.T) {
	g := initializeGame(t)

	g.State = game.GameStateLobby

	payload, err := BuildGameStatePayload(g)
	if err != nil {
		t.Fatalf("could not build game state payload %v", err)
	}

	if payload.State != "lobby" {
		t.Errorf("expected game state to be lobby, got %s", payload.State)
	}

	playersPayload := len(payload.Players)
	teamsPayload := len(payload.Teams)

	if playersPayload != len(g.Players) {
		t.Errorf("expected same amount of players, got %d and %d", playersPayload, len(g.Players))
	}

	if teamsPayload != len(g.Teams) {
		t.Errorf("expected same amount of teams, got %d and %d", teamsPayload, len(g.Teams))
	}

	if payload.Question != nil {
		t.Errorf("expected payload question to be nil in lobby")
	}

	if payload.Challenge != nil {
		t.Errorf("expected payload challenge to be nil in lobby")
	}
}

func TestBuildGameStatePayload_ChallengeRound(t *testing.T) {
	g := initializeGame(t)

	g.CurrentRound = 3
	g.CurrentItem = 0

	round, err := g.GetCurrentRound()
	if err != nil {
		t.Fatalf("could not get current round %v", err)
	}

	if round.Type != game.RoundTypeChallenge {
		t.Fatalf("expected round type to be challenge")
	}

	if len(round.Challenges) < 1 {
		t.Fatalf("expected atleast one challenge")
	}
	payload, err := BuildGameStatePayload(g)
	if err != nil {
		t.Fatalf("could not build game state payload %v", err)
	}

	challenge, err := g.GetCurrentChallenge()
	if err != nil {
		t.Fatalf("could not get current challenge")
	}

	if payload.State != "in_progress" {
		t.Errorf("expected game to be in progress")
	}

	if payload.CurrentRound != 3 {
		t.Errorf("expected current round to be 3, got %d", payload.CurrentRound)
	}
	if payload.CurrentItem != 0 {
		t.Errorf("expected current item to be 0, got %d", payload.CurrentItem)
	}

	if payload.RoundType != "challenge" {
		t.Errorf("expected round type challenge")
	}

	if payload.Challenge == nil {
		t.Fatalf("expected challenge to have a value")
	}

	if payload.Question != nil {
		t.Errorf("expected question to be nil")
	}

	if payload.Challenge.Title != challenge.Title {
		t.Errorf("expected challenge title to be the same as actual challenge title")
	}

	if payload.Challenge.Description != challenge.Description {
		t.Errorf("expected challenge description to be the same as actual challenge description")
	}
}

func TestBuildGameStateUpdateMessage(t *testing.T) {
	g := initializeGame(t)

	msgData, err := BuildGameStateUpdateMessage(g)
	if err != nil {
		t.Fatalf("could not update message %v", err)
	}

	var message Message

	err = json.Unmarshal(msgData, &message)
	if err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	if message.Type != GameStateUpdate {
		t.Errorf("expected message type to be %s", GameStateUpdate)
	}

	var gameStatePayload GameStatePayload
	err = json.Unmarshal(message.Payload, &gameStatePayload)
	if err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	round, err := g.GetCurrentRound()
	if err != nil {
		t.Fatalf("could not get current round %v", err)
	}

	question, err := g.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question %v", err)
	}

	if gameStatePayload.State != string(g.State) {
		t.Errorf("expected game state to be %s, got %s", string(g.State), gameStatePayload.State)
	}

	if gameStatePayload.RoundType != string(round.Type) {
		t.Errorf("expected round type to be %s, got %s", string(round.Type), gameStatePayload.RoundType)
	}

	if gameStatePayload.Question == nil {
		t.Fatalf("expected question to be in game state payload")
	}

	if gameStatePayload.Question.Question != question.Question {
		t.Errorf("expected question to be %s", question.Question)
	}
}

func TestHub_BroadcastGameState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	g := initializeGame(t)

	hub := NewHub()
	ch := make(chan error, 1)

	go func() {
		ch <- hub.BroadcastGameState(g)
	}()

	var msgData []byte

	select {
	case msgData = <-hub.broadcast:
		if err := <-ch; err != nil {
			t.Fatalf("could not broadcast game state %v", err)
		}
	case err := <-ch:
		if err != nil {
			t.Fatalf("could not broadcast game state %v", err)
		}
		t.Fatal("broadcast returned without sending message ")
	case <-ctx.Done():
		t.Fatalf("timed out waiting for broadcast message")
	}

	var msg Message
	err := json.Unmarshal(msgData, &msg)
	if err != nil {
		t.Fatalf("could not unmarshal data %v", err)
	}

	if msg.Type != GameStateUpdate {
		t.Errorf("expected message type %s, got %s", GameStateUpdate, msg.Type)
	}

	var payload GameStatePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Errorf("could not unarshal game state payload %v", err)
	}

	if payload.State != string(g.State) {
		t.Errorf("expected game state %s, got %s", g.State, payload.State)
	}

	if payload.Question == nil {
		t.Fatal("expected question in payload")
	}
}

func TestHub_RunBroadcastMessageToClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	g := initializeGame(t)

	hub := NewHub()

	go func() {
		hub.Run()
	}()

	client := Client{
		hub:  hub,
		game: g,
		send: make(chan []byte, 1),
	}

	hub.register <- &client

	if err := hub.BroadcastGameState(g); err != nil {
		t.Fatalf("could not broadcast game state %v", err)
	}

	var msgData []byte

	select {
	case msgData = <-client.send:
		//
	case <-ctx.Done():
		t.Fatalf("timed out waiting for broadcast message")
	}

	var msg Message
	if err := json.Unmarshal(msgData, &msg); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	if msg.Type != GameStateUpdate {
		t.Errorf("expected message type %s got %s", GameStateUpdate, msg.Type)
	}

	var payload GameStatePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	if payload.Question == nil {
		t.Fatal("expeced question in gamestate payload")
	}
}

func TestHandleBuzz_RegisterBuzz(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	g := initializeGame(t)

	g.CurrentRound = 2

	round, err := g.GetCurrentRound()
	if err != nil {
		t.Fatalf("expected to get current round got err %v", err)
	}

	if round.Type != game.RoundTypeFastestBuzz {
		t.Fatalf("expected round type %s, got %s", game.RoundTypeFastestBuzz, round.Type)
	}

	hub := NewHub()

	go func() {
		hub.Run()
	}()

	client := Client{
		hub:      hub,
		game:     g,
		playerID: 1,
		send:     make(chan []byte, 1),
	}

	hub.register <- &client

	client.handleBuzz()

	if g.BuzzedPlayerID == nil {
		t.Fatal("expected buzzed player id to be set")
	}
	if *g.BuzzedPlayerID != client.playerID {
		t.Errorf("expected to get player id %d got %d", client.playerID, *g.BuzzedPlayerID)
	}

	var msgData []byte

	select {
	case msgData = <-client.send:

	case <-ctx.Done():
		t.Fatalf("timed out waiting for broadcast message")

	}

	var msg Message
	if err := json.Unmarshal(msgData, &msg); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	if msg.Type != GameStateUpdate {
		t.Fatalf("expected message type to be %s, got %s", GameStateUpdate, msg.Type)
	}

	var payload GameStatePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}
	round, err = g.GetCurrentRound()
	if err != nil {
		t.Fatalf("could not get current round %v", err)
	}

	if payload.RoundType != string(round.Type) {
		t.Errorf("expected round type to be %s, got %s", round.Type, payload.RoundType)
	}

	if payload.BuzzedPlayerID == nil {
		t.Fatal("expected buzzedplayer id to be in payload")
	}
	if *payload.BuzzedPlayerID != client.playerID {
		t.Errorf("expected to get player id %d got %d", client.playerID, *payload.BuzzedPlayerID)
	}
}

func TestHandleAnswer_CorrectAnswer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	g := initializeGame(t)

	g.CurrentRound = 1
	g.CurrentItem = 0

	playerID := uint(1)
	var playerTeamID *uint
	for i := range g.Players {
		if g.Players[i].PlayerID == playerID {
			playerTeamID = g.Players[i].TeamID
			break
		}
	}
	if playerTeamID == nil {
		t.Fatalf("player team id was not found")
	}

	pointsBefore := 0
	for i := range g.Teams {
		if g.Teams[i].TeamID == *playerTeamID {
			pointsBefore = g.Teams[i].Points
		}
	}

	question, err := g.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question %v", err)
	}

	hub := NewHub()

	go func() {
		hub.Run()
	}()

	client := Client{
		hub:      hub,
		game:     g,
		playerID: 1,
		send:     make(chan []byte, 2),
	}

	hub.register <- &client

	answerPayload := AnswerPayload{
		Answer: question.CorrectAnswer,
	}
	payload, err := json.Marshal(answerPayload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}

	message := Message{
		Type:    Answer,
		Payload: payload,
	}

	client.handleAnswer(message)

	pointsAfter := 0
	for i := range g.Teams {
		if g.Teams[i].TeamID == *playerTeamID {
			pointsAfter = g.Teams[i].Points
			break
		}
	}

	if pointsAfter != pointsBefore+100 {
		t.Errorf("expected points to be %d, got %d", pointsBefore+100, pointsAfter)
	}

	if g.CurrentItem != 0 {
		t.Errorf("expected to remain on current item %d, got %d", 0, g.CurrentItem)
	}

	if g.CurrentRound != 1 {
		t.Errorf("expected to stay in round number %d got %d", 1, g.CurrentRound)
	}

	var msgData []byte

	select {
	case msgData = <-client.send:

	case <-ctx.Done():
		t.Fatalf("timed out waiting for broadcast message")
	}

	var resultMessage Message
	if err := json.Unmarshal(msgData, &resultMessage); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	if resultMessage.Type != AnswerResult {
		t.Fatalf("expected message type %s, got %s", AnswerResult, resultMessage.Type)
	}

	var resultPayload AnswerResultPayload
	if err := json.Unmarshal(resultMessage.Payload, &resultPayload); err != nil {
		t.Fatalf("could not unmarshal result payload %v", err)
	}

	if !resultPayload.IsCorrect {
		t.Error("expected answer to be correct")
	}

	if resultPayload.PlayerID != playerID {
		t.Errorf("expected player id %d, got %d", playerID, resultPayload.PlayerID)
	}

	if resultPayload.CorrectAnswer != question.CorrectAnswer {
		t.Errorf("expected correct answer %q, got %q", question.CorrectAnswer, resultPayload.CorrectAnswer)
	}

	select {
	case msgData = <-client.send:
	case <-ctx.Done():
		t.Fatal("timed out waiting for game state update")
	}

	var stateMessage Message
	if err := json.Unmarshal(msgData, &stateMessage); err != nil {
		t.Fatalf("could not unmarshal game state message %v", err)
	}

	if stateMessage.Type != GameStateUpdate {
		t.Errorf("expected message type %s, got %s", GameStateUpdate, stateMessage.Type)
	}
	var gameStatePayload GameStatePayload

	if err := json.Unmarshal(stateMessage.Payload, &gameStatePayload); err != nil {
		t.Fatalf("could not unmarshal game state payload: %v", err)
	}

	if gameStatePayload.Question == nil {
		t.Fatal("expected question in game state payload")
	}

	if gameStatePayload.CurrentItem != 0 {
		t.Errorf("expected current item to remain at 0, got %d", gameStatePayload.CurrentItem)
	}
}

func TestHandleAnswer_WrongAnswer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	g := initializeGame(t)

	g.CurrentRound = 1
	g.CurrentItem = 0

	playerID := uint(1)
	var playerTeamID *uint
	for i := range g.Players {
		if g.Players[i].PlayerID == playerID {
			playerTeamID = g.Players[i].TeamID
			break
		}
	}
	if playerTeamID == nil {
		t.Fatalf("player team id was not found")
	}

	pointsBefore := 0
	for i := range g.Teams {
		if g.Teams[i].TeamID == *playerTeamID {
			pointsBefore = g.Teams[i].Points
		}
	}

	question, err := g.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question %v", err)
	}

	wrongAnswer := "wrong answer"
	if wrongAnswer == question.CorrectAnswer {
		wrongAnswer = "another wrong answer"
	}
	hub := NewHub()

	go func() {
		hub.Run()
	}()

	client := Client{
		hub:      hub,
		game:     g,
		playerID: 1,
		send:     make(chan []byte, 2),
	}

	hub.register <- &client

	answerPayload := AnswerPayload{
		Answer: wrongAnswer,
	}
	payload, err := json.Marshal(answerPayload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}

	message := Message{
		Type:    Answer,
		Payload: payload,
	}

	client.handleAnswer(message)

	nextQuestion, err := g.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get next question %v", err)
	}

	pointsAfter := 0
	for i := range g.Teams {
		if g.Teams[i].TeamID == *playerTeamID {
			pointsAfter = g.Teams[i].Points
			break
		}
	}

	if pointsAfter != pointsBefore {
		t.Errorf("expected points to be %d, got %d", pointsBefore, pointsAfter)
	}

	if g.CurrentItem != 0 {
		t.Errorf("expected to go remain at item %d, got %d", 0, g.CurrentItem)
	}

	if g.CurrentRound != 1 {
		t.Errorf("expected to stay in round number %d got %d", 1, g.CurrentRound)
	}

	var msgData []byte

	select {
	case msgData = <-client.send:

	case <-ctx.Done():
		t.Fatalf("timed out waiting for answer result")
	}

	var resultMessage Message
	if err := json.Unmarshal(msgData, &resultMessage); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	if resultMessage.Type != AnswerResult {
		t.Fatalf("expected message type %s, got %s", AnswerResult, resultMessage.Type)
	}

	var resultPayload AnswerResultPayload
	if err := json.Unmarshal(resultMessage.Payload, &resultPayload); err != nil {
		t.Fatalf("could not unmarshal result payload %v", err)
	}

	if resultPayload.IsCorrect {
		t.Error("expected answer to be incorrect")
	}

	if resultPayload.PlayerID != playerID {
		t.Errorf("expected player id %d, got %d", playerID, resultPayload.PlayerID)
	}

	if resultPayload.CorrectAnswer != question.CorrectAnswer {
		t.Errorf("expected correct answer %q, got %q", question.CorrectAnswer, resultPayload.CorrectAnswer)
	}

	select {
	case msgData = <-client.send:
	case <-ctx.Done():
		t.Fatal("timed out waiting for game state update")
	}
	var stateMessage Message
	if err := json.Unmarshal(msgData, &stateMessage); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}
	if stateMessage.Type != GameStateUpdate {
		t.Errorf("expected message type %s, got %s", GameStateUpdate, stateMessage.Type)
	}
	var gameStatePayload GameStatePayload

	if err := json.Unmarshal(stateMessage.Payload, &gameStatePayload); err != nil {
		t.Fatalf("could not unmarshal game state payload: %v", err)
	}

	if gameStatePayload.Question.Question != nextQuestion.Question {
		t.Errorf("expected to get next question %q, got %q", nextQuestion.Question, gameStatePayload.Question.Question)
	}

	if gameStatePayload.Question == nil {
		t.Fatal("expected question in game state payload")
	}
}

func TestJoinGame_FirstPlayerBecomesHost(t *testing.T) {
	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("could not load rounds: %v", err)
	}

	g := game.NewGame(rounds)
	hub := NewHub()

	go func() {
		hub.Run()
	}()

	client := Client{
		hub:  hub,
		game: g,
		send: make(chan []byte, 1),
	}

	joinPayload := game.JoinGamePayload{
		PlayerName: "John",
	}

	payload, err := json.Marshal(joinPayload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}
	msgJoinGame := Message{
		Type:    JoinGame,
		Payload: payload,
	}
	client.joinGame(msgJoinGame)

	if client.playerID == 0 {
		t.Fatal("expected client to recieve a player id")
	}
	if g.HostPlayerID != client.playerID {
		t.Errorf("expected host player id to be %d, got %d", client.playerID, g.HostPlayerID)
	}
}

func TestJoinGame_SecondPlayerDoesntReplaceHost(t *testing.T) {
	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("could not load rounds: %v", err)
	}

	g := game.NewGame(rounds)
	hub := NewHub()

	go func() {
		hub.Run()
	}()

	client := Client{
		hub:  hub,
		game: g,
		send: make(chan []byte, 1),
	}

	joinPayload := game.JoinGamePayload{
		PlayerName: "John",
	}

	payload, err := json.Marshal(joinPayload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}
	msgJoinGame := Message{
		Type:    JoinGame,
		Payload: payload,
	}
	client.joinGame(msgJoinGame)

	hostID := g.HostPlayerID

	if hostID == 0 {
		t.Fatal("expected host to have host id")
	}

	newClient := Client{
		hub:  hub,
		game: g,
		send: make(chan []byte, 1),
	}

	newJoinPayload := game.JoinGamePayload{
		PlayerName: "Doe",
	}

	newClientPayload, err := json.Marshal(newJoinPayload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}
	msgJoinGame = Message{
		Type:    JoinGame,
		Payload: newClientPayload,
	}

	newClient.joinGame(msgJoinGame)

	if newClient.playerID == 0 {
		t.Fatal("expected new client to have player id ")
	}
	if newClient.playerID == client.playerID {
		t.Errorf("expected different player ids, but both were %d", client.playerID)
	}
	if g.HostPlayerID != hostID {
		t.Errorf("expected host player id to be %d, got %d", hostID, g.HostPlayerID)
	}

	if len(g.Players) != 2 {
		t.Errorf("expected game to have 2 players, got %d", len(g.Players))
	}
}

func TestHandleStartGame_HostStartsGame(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("could not load rounds: %v", err)
	}

	hub := NewHub()

	go hub.Run()
	g := game.NewGame(rounds)
	host := g.JoinGame("John")
	g.JoinGame("Doe")

	// John er host
	g.HostPlayerID = host.PlayerID

	client := Client{
		hub:      hub,
		game:     g,
		playerID: host.PlayerID,
		send:     make(chan []byte, 1),
	}

	hub.register <- &client
	client.handleStartGame()

	if g.State != game.GameStateInProgress {
		t.Fatalf("expected game state to be %s, got %s", game.GameStateInProgress, g.State)
	}

	var msgData []byte

	select {
	case msgData = <-client.send:

	case <-ctx.Done():
		t.Fatalf("timed out waiting for broadcast message")
	}

	var msg Message
	if err := json.Unmarshal(msgData, &msg); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	if msg.Type != GameStateUpdate {
		t.Errorf("expected message type %s got %s", GameStateUpdate, msg.Type)
	}

	var payload GameStatePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}

	question, err := g.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question %v", err)
	}

	if payload.Question.Question != question.Question {
		t.Errorf("expected payload to contain the first question %s", question.Question)
	}
}

func TestHandleStartGame_NonHostCantStartGame(t *testing.T) {
	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("could not load rounds: %v", err)
	}

	hub := NewHub()

	go hub.Run()
	g := game.NewGame(rounds)
	john := g.JoinGame("John")
	doe := g.JoinGame("Doe")

	// John er host
	g.HostPlayerID = john.PlayerID

	doeClient := Client{
		hub:      hub,
		game:     g,
		playerID: doe.PlayerID,
		send:     make(chan []byte, 1),
	}
	hub.register <- &doeClient
	doeClient.handleStartGame()

	if g.State != game.GameStateLobby {
		t.Fatalf("expected game state to be %s, got %s", game.GameStateLobby, g.State)
	}

	if g.HostPlayerID != john.PlayerID {
		t.Errorf("expected player %d to be host player got %d", john.PlayerID, g.HostPlayerID)
	}

	if doe.PlayerID == john.PlayerID {
		t.Errorf("expected different player ids for player %d and player %d", doe.PlayerID, john.PlayerID)
	}
}

func TestJoinGame_BroadcastUpdatedLobbyState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("could not load rounds: %v", err)
	}

	g := game.NewGame(rounds)

	hub := NewHub()

	go hub.Run()

	john := Client{
		hub:  hub,
		game: g,
		send: make(chan []byte, 2),
	}

	hub.register <- &john

	joinPayload := game.JoinGamePayload{
		PlayerName: "John",
	}

	payloadData, err := json.Marshal(joinPayload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}
	msgJoinGame := Message{
		Type:    JoinGame,
		Payload: payloadData,
	}
	john.joinGame(msgJoinGame)

	if john.playerID == 0 {
		t.Fatal("expected client to recieve a player id")
	}
	receivedPlayerJoined := false
	receivedGameState := false
	var gameStatePayload GameStatePayload

	for !receivedGameState {
		select {
		case msgData := <-john.send:
			var msg Message
			if err := json.Unmarshal(msgData, &msg); err != nil {
				t.Fatalf("could not unmarshal message: %v", err)
			}

			switch msg.Type {
			case PlayerJoined:
				receivedPlayerJoined = true

			case GameStateUpdate:
				if err := json.Unmarshal(msg.Payload, &gameStatePayload); err != nil {
					t.Fatalf("could not unmarshal game state payload: %v", err)
				}
				receivedGameState = true

			default:
				t.Fatalf("unexpected message type: %s", msg.Type)
			}

		case <-ctx.Done():
			t.Fatal("timed out waiting for lobby messages")
		}
	}

	if !receivedPlayerJoined {
		t.Error("expected to receive PlayerJoined")
	}

	if gameStatePayload.State != string(game.GameStateLobby) {
		t.Errorf("expected game state %s, got %s",
			game.GameStateLobby, gameStatePayload.State)
	}

	foundJohn := false
	for i := range gameStatePayload.Players {
		if gameStatePayload.Players[i].PlayerID == john.playerID {
			foundJohn = true
			break
		}
	}

	if !foundJohn {
		t.Error("expected John to be present in the player list")
	}

	if gameStatePayload.HostPlayerID != john.playerID {
		t.Errorf("expected John to be host with ID %d, got %d",
			john.playerID, gameStatePayload.HostPlayerID)
	}

	if gameStatePayload.Question != nil {
		t.Error("expected question to be nil in lobby")
	}

	if gameStatePayload.Challenge != nil {
		t.Error("expected challenge to be nil in lobby")
	}
}

func TestWebSocket_JoinGame(t *testing.T) {
	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("could not load rounds: %v", err)
	}

	g := game.NewGame(rounds)

	hub := NewHub()

	go hub.Run()

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, g, w, r)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	header := http.Header{}
	header.Set("Origin", "http://localhost:5173")
	conn, _, err := gorilla.DefaultDialer.Dial(wsURL+"/ws", header)
	if err != nil {
		t.Fatalf("could not connect to websocket %v", err)
	}
	defer conn.Close()

	joinPayload := game.JoinGamePayload{
		PlayerName: "John",
	}

	payloadData, err := json.Marshal(joinPayload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}
	msgJoinGame := Message{
		Type:    JoinGame,
		Payload: payloadData,
	}

	if err := conn.WriteJSON(msgJoinGame); err != nil {
		t.Fatalf("could not write json %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("could not set read deadline: %v", err)
	}

	var playerJoinedMessage Message

	if err := conn.ReadJSON(&playerJoinedMessage); err != nil {
		t.Fatalf("could not read PlayerJoined message: %v", err)
	}

	if playerJoinedMessage.Type != PlayerJoined {
		t.Fatalf("expected message type %s, got %s", PlayerJoined, playerJoinedMessage.Type)
	}

	var gameStateMessage Message

	if err := conn.ReadJSON(&gameStateMessage); err != nil {
		t.Fatalf("could not read GameStateUpdate message: %v", err)
	}

	if gameStateMessage.Type != GameStateUpdate {
		t.Fatalf("expected message type %s, got %s", GameStateUpdate, gameStateMessage.Type)
	}

	var gameStatePayload GameStatePayload

	if err := json.Unmarshal(gameStateMessage.Payload, &gameStatePayload); err != nil {
		t.Fatalf("could not unmarshal game state payload: %v", err)
	}

	if gameStatePayload.State != string(game.GameStateLobby) {
		t.Errorf("expected state %s, got %s", game.GameStateLobby, gameStatePayload.State)
	}

	if gameStatePayload.HostPlayerID == 0 {
		t.Error("expected a host to be assigned")
	}

	if gameStatePayload.Question != nil {
		t.Error("expected question to be nil in lobby")
	}

	if gameStatePayload.Challenge != nil {
		t.Error("expected challenge to be nil in lobby")
	}
}

func TestWebSocket_HostStartsGame(t *testing.T) {
	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("could not load rounds: %v", err)
	}

	g := game.NewGame(rounds)
	hub := NewHub()

	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, g, w, r)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	header := http.Header{}
	header.Set("Origin", "http://localhost:5173")

	// Helper til at læse en enkelt besked fra en WebSocket.
	readMessage := func(conn *gorilla.Conn) Message {
		t.Helper()

		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatalf("could not set read deadline: %v", err)
		}

		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("could not read message: %v", err)
		}

		return msg
	}

	connJohn, _, err := gorilla.DefaultDialer.Dial(wsURL+"/ws", header)
	if err != nil {
		t.Fatalf("could not connect John to websocket: %v", err)
	}
	defer connJohn.Close()

	johnJoinPayload := game.JoinGamePayload{
		PlayerName: "John",
	}

	johnPayloadData, err := json.Marshal(johnJoinPayload)
	if err != nil {
		t.Fatalf("could not marshal John's payload: %v", err)
	}

	johnMsgJoinGame := Message{
		Type:    JoinGame,
		Payload: johnPayloadData,
	}

	if err := connJohn.WriteJSON(johnMsgJoinGame); err != nil {
		t.Fatalf("could not send John's join message: %v", err)
	}

	johnJoinedMessage := readMessage(connJohn)

	if johnJoinedMessage.Type != PlayerJoined {
		t.Fatalf(
			"expected message type %s, got %s",
			PlayerJoined,
			johnJoinedMessage.Type,
		)
	}

	var johnJoined game.PlayerJoinedPayload
	if err := json.Unmarshal(johnJoinedMessage.Payload, &johnJoined); err != nil {
		t.Fatalf("could not unmarshal John's joined payload: %v", err)
	}

	if johnJoined.PlayerName != "John" {
		t.Fatalf("expected joined player John, got %s", johnJoined.PlayerName)
	}

	johnLobbyMessage := readMessage(connJohn)

	if johnLobbyMessage.Type != GameStateUpdate {
		t.Fatalf(
			"expected message type %s, got %s",
			GameStateUpdate,
			johnLobbyMessage.Type,
		)
	}

	var johnLobbyState GameStatePayload
	if err := json.Unmarshal(johnLobbyMessage.Payload, &johnLobbyState); err != nil {
		t.Fatalf("could not unmarshal John's lobby state: %v", err)
	}

	if johnLobbyState.State != string(game.GameStateLobby) {
		t.Fatalf("expected lobby state, got %s", johnLobbyState.State)
	}

	if johnLobbyState.HostPlayerID != johnJoined.PlayerID {
		t.Fatalf(
			"expected John (ID %d) to be host, got %d",
			johnJoined.PlayerID,
			johnLobbyState.HostPlayerID,
		)
	}

	connDoe, _, err := gorilla.DefaultDialer.Dial(wsURL+"/ws", header)
	if err != nil {
		t.Fatalf("could not connect Doe to websocket: %v", err)
	}
	defer connDoe.Close()

	doeJoinPayload := game.JoinGamePayload{
		PlayerName: "Doe",
	}

	doePayloadData, err := json.Marshal(doeJoinPayload)
	if err != nil {
		t.Fatalf("could not marshal Doe's payload: %v", err)
	}

	doeMsgJoinGame := Message{
		Type:    JoinGame,
		Payload: doePayloadData,
	}

	if err := connDoe.WriteJSON(doeMsgJoinGame); err != nil {
		t.Fatalf("could not send Doe's join message: %v", err)
	}

	doeJoinedMessage := readMessage(connJohn)

	if doeJoinedMessage.Type != PlayerJoined {
		t.Fatalf(
			"expected message type %s, got %s",
			PlayerJoined,
			doeJoinedMessage.Type,
		)
	}

	var doeJoined game.PlayerJoinedPayload
	if err := json.Unmarshal(doeJoinedMessage.Payload, &doeJoined); err != nil {
		t.Fatalf("could not unmarshal Doe's joined payload: %v", err)
	}

	if doeJoined.PlayerName != "Doe" {
		t.Fatalf("expected joined player Doe, got %s", doeJoined.PlayerName)
	}

	doeLobbyMessage := readMessage(connJohn)

	if doeLobbyMessage.Type != GameStateUpdate {
		t.Fatalf(
			"expected message type %s, got %s",
			GameStateUpdate,
			doeLobbyMessage.Type,
		)
	}

	var lobbyStateWithBothPlayers GameStatePayload
	if err := json.Unmarshal(doeLobbyMessage.Payload, &lobbyStateWithBothPlayers); err != nil {
		t.Fatalf("could not unmarshal lobby state: %v", err)
	}

	if len(lobbyStateWithBothPlayers.Players) != 2 {
		t.Fatalf(
			"expected 2 players in lobby, got %d",
			len(lobbyStateWithBothPlayers.Players),
		)
	}

	if lobbyStateWithBothPlayers.HostPlayerID != johnJoined.PlayerID {
		t.Fatalf(
			"expected John to remain host with ID %d, got %d",
			johnJoined.PlayerID,
			lobbyStateWithBothPlayers.HostPlayerID,
		)
	}

	startMessage := Message{
		Type: StartGame,
	}

	if err := connJohn.WriteJSON(startMessage); err != nil {
		t.Fatalf("could not send start game message: %v", err)
	}

	johnPayload := readGameInProgress(t, connJohn)
	doePayload := readGameInProgress(t, connDoe)

	if johnPayload.State != string(game.GameStateInProgress) {
		t.Errorf("expected John to receive in_progress, got %s", johnPayload.State)
	}

	if doePayload.State != string(game.GameStateInProgress) {
		t.Errorf("expected Doe to receive in_progress, got %s", doePayload.State)
	}

	if johnPayload.Question == nil {
		t.Error("expected John to receive the first question")
	}

	if doePayload.Question == nil {
		t.Error("expected Doe to receive the first question")
	}
}

func TestHandleCompleteChallenge_NonHostCannotCompleteChallenge(t *testing.T) {
	g := initializeGame(t)

	g.CurrentRound = 3
	g.CurrentItem = 0

	hub := NewHub()

	go hub.Run()

	pointsBefore := make(map[uint]int)

	for _, team := range g.Teams {
		pointsBefore[team.TeamID] = team.Points
	}

	hostPlayerID := g.Players[0].PlayerID
	playerID := g.Players[1].PlayerID
	g.HostPlayerID = hostPlayerID
	client := Client{
		hub:      hub,
		game:     g,
		send:     make(chan []byte, 1),
		playerID: playerID,
	}

	hub.register <- &client

	team1 := g.Teams[0].TeamID
	team2 := g.Teams[1].TeamID
	teamPoints := map[uint]int{
		team1: 100,
		team2: 0,
	}

	payload := CompleteChallengePayload{
		teamPoints,
	}

	p, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}
	if client.playerID == g.HostPlayerID {
		t.Fatal("test setup error: client is the host")
	}
	msg := Message{
		Type:    CompleteChallenge,
		Payload: p,
	}

	client.handleCompleteChallenge(msg)

	for _, team := range g.Teams {
		expectedPoints := pointsBefore[team.TeamID]
		if team.Points != expectedPoints {
			t.Errorf("team %d: expected points to remain %d, got %d", team.TeamID, expectedPoints, team.Points)
		}
	}

	if g.CurrentItem != 0 {
		t.Errorf("expected to stay at current item %d, got %d", 0, g.CurrentItem)
	}

	select {
	case msgData := <-client.send:
		t.Fatalf("did not expect broadcast, got %s", string(msgData))
	default:

	}

	if g.HostPlayerID != hostPlayerID {
		t.Errorf("expected host player to have id %d, got %d", hostPlayerID, g.HostPlayerID)
	}
}

func TestHandleCompleteChallenge_MissingTeamPoints(t *testing.T) {
	g := initializeGame(t)

	g.CurrentRound = 3
	g.CurrentItem = 0

	hub := NewHub()

	go hub.Run()

	pointsBefore := make(map[uint]int)

	for _, team := range g.Teams {
		pointsBefore[team.TeamID] = team.Points
	}

	playerID := g.Players[0].PlayerID
	g.HostPlayerID = playerID
	client := Client{
		hub:      hub,
		game:     g,
		send:     make(chan []byte, 1),
		playerID: playerID,
	}

	hub.register <- &client

	team1 := g.Teams[0].TeamID
	teamPoints := map[uint]int{
		team1: 100,
	}

	payload := CompleteChallengePayload{
		teamPoints,
	}

	p, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}

	msg := Message{
		Type:    CompleteChallenge,
		Payload: p,
	}

	client.handleCompleteChallenge(msg)

	for _, team := range g.Teams {
		expectedPoints := pointsBefore[team.TeamID]
		if team.Points != expectedPoints {
			t.Errorf("team %d: expected points to remain %d, got %d", team.TeamID, expectedPoints, team.Points)
		}
	}
	if g.CurrentItem != 0 {
		t.Errorf("expected to stay at current item %d, got %d", 0, g.CurrentItem)
	}

	select {
	case msgData := <-client.send:
		t.Fatalf("did not expect broadcast, got %s", string(msgData))
	default:

	}
}

func TestHandleCompleteChallenge_UnknownTeamID(t *testing.T) {
	g := initializeGame(t)

	g.CurrentRound = 3
	g.CurrentItem = 0

	hub := NewHub()

	go hub.Run()

	pointsBefore := make(map[uint]int)

	for _, team := range g.Teams {
		pointsBefore[team.TeamID] = team.Points
	}

	playerID := g.Players[0].PlayerID
	g.HostPlayerID = playerID
	client := Client{
		hub:      hub,
		game:     g,
		send:     make(chan []byte, 1),
		playerID: playerID,
	}

	hub.register <- &client

	team1 := g.Teams[0].TeamID
	team2 := g.Teams[1].TeamID
	unknownTeamID := uint(0)
	for _, team := range g.Teams {
		if team.TeamID >= unknownTeamID {
			unknownTeamID = team.TeamID + 1
		}
	}
	teamPoints := map[uint]int{
		team1:         100,
		team2:         0,
		unknownTeamID: 100,
	}

	payload := CompleteChallengePayload{
		teamPoints,
	}

	p, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}

	msg := Message{
		Type:    CompleteChallenge,
		Payload: p,
	}

	client.handleCompleteChallenge(msg)

	for _, team := range g.Teams {
		expectedPoints := pointsBefore[team.TeamID]
		if team.Points != expectedPoints {
			t.Errorf("team %d: expected points to remain %d, got %d", team.TeamID, expectedPoints, team.Points)
		}
	}
	if g.CurrentItem != 0 {
		t.Errorf("expected to stay at current item %d, got %d", 0, g.CurrentItem)
	}

	select {
	case msgData := <-client.send:
		t.Fatalf("did not expect broadcast, got %s", string(msgData))
	default:

	}
}
