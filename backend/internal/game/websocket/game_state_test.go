package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/danielroschmann/interactive-quiz-game/backend/internal/game"
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
		send:     make(chan []byte, 1),
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

	if g.CurrentItem != 1 {
		t.Errorf("expected to go on to next item %d, got %d", 1, g.CurrentItem)
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

	var msg Message
	if err := json.Unmarshal(msgData, &msg); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}
	if msg.Type != GameStateUpdate {
		t.Errorf("expected message type %s, got %s", GameStateUpdate, msg.Type)
	}
	var gameStatePayload GameStatePayload

	if err := json.Unmarshal(msg.Payload, &gameStatePayload); err != nil {
		t.Fatalf("could not unmarshal game state payload: %v", err)
	}

	if gameStatePayload.Question == nil {
		t.Fatal("expected question in game state payload")
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
		send:     make(chan []byte, 1),
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

	if g.CurrentItem != 1 {
		t.Errorf("expected to go on to next item %d, got %d", 1, g.CurrentItem)
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

	var msg Message
	if err := json.Unmarshal(msgData, &msg); err != nil {
		t.Fatalf("could not unmarshal message %v", err)
	}
	if msg.Type != GameStateUpdate {
		t.Errorf("expected message type %s, got %s", GameStateUpdate, msg.Type)
	}
	var gameStatePayload GameStatePayload

	if err := json.Unmarshal(msg.Payload, &gameStatePayload); err != nil {
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
