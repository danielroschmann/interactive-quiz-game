import { useRef, useState } from 'react'
import './App.css'



type Player = {
  player_id: number,
  player_name: string,
  team_id?: number
}

type GameState = 'lobby' | 'in_progress' | 'finished'
type GameStatePayload = {
  state: GameState
  players: Player[]
  host_player_id: number
  current_round: number
  current_item: number
  round_type: string
}

function App() {

  const [username, setUsername] = useState('')
  const [hostPlayerId, setHostPlayerId] = useState<number | null>(null)
  const [status, setStatus] = useState('Disconnected')
  const socketRef = useRef<WebSocket | null>(null)
  const [players, setPlayers] = useState<Player[]>([])
  const [playerId, setPlayerId] = useState<number | null>(null)
  const [gameState, setGameState] = useState<GameState>('lobby')
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
