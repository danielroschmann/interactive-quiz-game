import { type FormEvent, useRef, useState } from 'react'
import './App.css'



type Player = {
  player_id: number,
  player_name: string,
  team_id?: number
}

type Question = {
  question: string
  answer_options?: string[]
}

type GameState = 'lobby' | 'in_progress' | 'finished'
type RoundType = 'multiple_choice' | 'fastest_buzz' | 'challenge'

type GameStatePayload = {
  state: GameState
  players: Player[]
  host_player_id: number
  current_round: number
  current_item: number
  round_type: RoundType
  question?: Question | null
  buzzed_player_id?: number
}

function App() {

  const [username, setUsername] = useState('')
  const [hostPlayerId, setHostPlayerId] = useState<number | null>(null)
  const [status, setStatus] = useState('Disconnected')
  const [roundType, setRoundType] = useState<RoundType>('multiple_choice')
  const [question, setQuestion] = useState<Question | null>(null)
  const [buzzPlayerId, setBuzzPlayerId] = useState<number | null>(null)
  const [buzzAnswer, setBuzzAnswer] = useState('')
  const socketRef = useRef<WebSocket | null>(null)
  const [players, setPlayers] = useState<Player[]>([])
  const [playerId, setPlayerId] = useState<number | null>(null)
  const [gameState, setGameState] = useState<GameState>('lobby')

  const buzzedPlayer = players.find(
    (player) => player.player_id === buzzPlayerId
  )
  function handleJoinGame() {
    const name = username.trim()


    if (!name) {
      setStatus('Enter your name')
      return
    }

    const socket = new WebSocket('ws://localhost:8080/ws')
    socketRef.current = socket

    socket.addEventListener('open', () => {
      setStatus('Connected')

      socket.send(JSON.stringify({
        type: 'join_game',
        payload: {
          player_name: name,
        },
      }))
    })

    socket.addEventListener('message', (event) => {
      const message = JSON.parse(event.data)
      if (message.type === 'player_joined') {
        const joinedPlayer = message.payload

        if (joinedPlayer.player_name === name) {
          setPlayerId(joinedPlayer.player_id)
        }
      }
      if (message.type == 'game_state_update') {
        const payload = message.payload
        setPlayers(payload.players)
        setHostPlayerId(payload.host_player_id)
        setGameState(payload.state)
        setQuestion(payload.question ?? null)
        setRoundType(payload.round_type)
        setBuzzPlayerId(payload.buzzed_player_id ?? null)
      }

      console.log('Message from server:', message)
    })

    socket.addEventListener('close', () => {
      setStatus('Disconnected')

    })

    socket.addEventListener('error', () => {
      setStatus('Connection error')
    })
  }

  function handleStartGame() {

    const socket = socketRef.current
    if (!socket || socket.readyState != WebSocket.OPEN) {
      setStatus('Not connected')
      return
    }

    socket.send(JSON.stringify({
      type: 'start_game'
    }))

  }

  function handleAnswer(answer: string) {
    const socket = socketRef.current

    if (!socket || socket.readyState !== WebSocket.OPEN) {
      setStatus('Not connected')
      return
    }

    socket.send(JSON.stringify({
      type: 'answer',
      payload: {
        answer: answer,
      },
    }))
  }

  function handleBuzz() {
    const socket = socketRef.current

    if (!socket || socket.readyState !== WebSocket.OPEN) {
      setStatus('Not connected')
      return
    }

    socket.send(JSON.stringify({
      type: 'buzz',
    }))
  }

  function handleBuzzAnswerSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const answer = buzzAnswer.trim()

    if (!answer) {
      return
    }

    handleAnswer(answer)
    setBuzzAnswer('')
  }


  return (


    <section id="center">
      {gameState === 'lobby' && (
        <>
          <h2>Players in Lobby</h2>
          <ul>
            {players.map((player) => (
              <li key={player.player_id}>
                {player.player_name}


                {player.player_id === hostPlayerId && (
                  <span> (Host)</span>
                )}
              </li>
            ))}
          </ul>
          <label>
            Name:
            <input
              type="text"
              name="username"
              className='counter'
              value={username}
              onChange={(event => setUsername(event.target.value))}>
            </input>
          </label>
          <button
            type="button"
            className="counter"
            onClick={handleJoinGame}
          >
            Join Game
          </button>
          {playerId !== null && playerId === hostPlayerId && (
            <button
              type="button"
              className='counter'
              onClick={handleStartGame}
            >
              Start Game
            </button>
          )}
        </>
      )}

      {gameState === 'in_progress' && (
        <>
          <h2>Quiz</h2>
          <p>Game is started!</p>
          {question ? (
            <>
              <h3>{question.question}</h3>

              {roundType === 'multiple_choice' && (
                <ul>
                  {(question.answer_options ?? []).map((answerOption, index) => (
                    <li key={`${index}-${answerOption}`}>
                      <button
                        type='button'
                        onClick={() => handleAnswer(answerOption)}
                      >
                        {answerOption}
                      </button>

                    </li>
                  ))}
                </ul>
              )}

              {roundType === 'fastest_buzz' && (
                <>
                  {buzzPlayerId === null && (
                    <button
                      type='button'
                      onClick={handleBuzz}
                    >
                      Buzz
                    </button>
                  )}

                  {buzzedPlayer && (
                    <h1>{buzzedPlayer.player_name} has buzzed first!</h1>
                  )}

                  {playerId !== null && buzzPlayerId !== null && playerId === buzzPlayerId && (
                    <form onSubmit={handleBuzzAnswerSubmit}>
                      <label>
                        Type Answer
                        <input
                          type="text"
                          name="buzz_answer"
                          value={buzzAnswer}
                          onChange={(event) => setBuzzAnswer(event.target.value)}
                        >
                        </input>
                      </label>

                      <button
                        type='submit'
                      >
                        Submit Answer
                      </button>
                    </form>
                  )}
                </>
              )}

            </>
          ) : (
            <p>Loading question...</p>
          )}
        </>
      )}

      {gameState === 'finished' && (
        <>
          <h2>Game Over</h2>
          <p>Thank you for playing</p>
        </>
      )}

    </section >
  )
}

export default App
