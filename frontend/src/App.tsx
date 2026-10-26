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
  null
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
      if (message.type == 'game_state_update') {
        const gameState = message.payload
        setPlayers(gameState.players)
        setHostPlayerId(gameState.host_player_id)
        console.log(gameState.players)
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
      </section >
    </>
  )
}

export default App
