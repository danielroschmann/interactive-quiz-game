import { useRef, useState } from 'react'
import './App.css'



type Player = {
  player_id: number,
  player_name: string,
  team_id?: number
}

function App() {

  const [username, setUsername] = useState('')
  const [hostPlayerId, setHostPlayerId] = useState<number | null>(null)
  const [status, setStatus] = useState('Disconnected')
  const socketRef = useRef<WebSocket | null>(null)
  const [players, setPlayers] = useState<Player[]>([])
  const [playerId, setPlayerId] = useState<number | null>(null)
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
        const gameState = message.payload
        setPlayers(gameState.players)
        setHostPlayerId(gameState.host_player_id)
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


    <>
      <section id="center">

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
      </section >
    </>
  )
}

export default App
