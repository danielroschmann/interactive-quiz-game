import { useRef, useState } from 'react'
import './App.css'




function App() {

  const [username, setUsername] = useState('')
  const [status, setStatus] = useState('Disconnected')
  const socketRef = useRef<WebSocket | null>(null)

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
      </section>
    </>
  )
}

export default App
