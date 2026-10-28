import { type FormEvent, useRef, useState } from 'react'
import './App.css'

type AnswerResult = {
  player_id: number,
  is_correct: boolean,
  correct_answer: string
}

type Player = {
  player_id: number,
  player_name: string,
  team_id?: number
}

type Team = {
  team_id: number,
  team_name: string,
  points: number
}

type Question = {
  question: string
  answer_options?: string[]
}

type Challenge = {
  title: string
  description: string
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
  challenge?: Challenge | null
  buzzed_player_id?: number
}

function App() {

  const [username, setUsername] = useState('')
  const [hostPlayerId, setHostPlayerId] = useState<number | null>(null)
  const [status, setStatus] = useState('Disconnected')
  const [roundType, setRoundType] = useState<RoundType>('multiple_choice')
  const [question, setQuestion] = useState<Question | null>(null)
  const [challenge, setChallenge] = useState<Challenge | null>(null)
  const [teams, setTeams] = useState<Team[]>([])
  const [buzzPlayerId, setBuzzPlayerId] = useState<number | null>(null)
  const [challengePoints, setChallengePoints] = useState<Record<number, string>>({})
  const [buzzAnswer, setBuzzAnswer] = useState('')
  const socketRef = useRef<WebSocket | null>(null)
  const [players, setPlayers] = useState<Player[]>([])
  const [playerId, setPlayerId] = useState<number | null>(null)
  const [gameState, setGameState] = useState<GameState>('lobby')
  const [answerResult, setAnswerResult] = useState<AnswerResult | null>(null)

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
      if (message.type == 'answer_result') {
        setAnswerResult(message.payload ?? null)
      }
      if (message.type == 'game_state_update') {
        const payload = message.payload
        setPlayers(payload.players)
        setTeams(payload.teams)
        setHostPlayerId(payload.host_player_id)
        setGameState(payload.state)
        setQuestion(payload.question ?? null)
        setRoundType(payload.round_type)
        setBuzzPlayerId(payload.buzzed_player_id ?? null)
        setChallenge(payload.challenge ?? null)
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

  function handleCompleteChallenge(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const socket = socketRef.current

    if (!socket || socket.readyState !== WebSocket.OPEN) {
      setStatus('Not connected')
      return
    }

    const teamPoints: Record<string, number> = {}
    for (const team of teams) {
      const input = challengePoints[team.team_id]

      if (input === undefined || input.trim() === '') {
        setStatus(`Enter points for ${team.team_name}`)
        return
      }

      const points = Number(input)

      if (!Number.isInteger(points) || points < 0) {
        setStatus(`Enter valid points for ${team.team_name}`)
        return
      }

      teamPoints[team.team_id] = points

    }

    socket.send(JSON.stringify({
      type: 'complete_challenge',
      payload: {
        team_points: teamPoints,
      },
    }))

  }


  return (


    <section id="center">
      <p>{status}</p>
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
          <div className='scoreboard'>
            <h3>Scoreboard</h3>
            <div className='teams'>
              {teams.map((team) =>
                <div className='team-score' key={team.team_id}>
                  <h3>{team.team_name}</h3>
                  <p>{team.points} points</p>
                </div>
              )}
            </div>
          </div>
          {answerResult && (
            <div className={answerResult.is_correct ? 'answer-result correct' : 'answer-result incorrect'}>
              <h3>
                {answerResult.is_correct ? 'Correct Answer!' : 'Incorrect Answer'}
              </h3>
              {!answerResult.is_correct && (
                <p>Correct answer was: {answerResult.correct_answer}</p>
              )}
            </div>
          )}
          {roundType === 'challenge' ? (
            challenge ? (
              <>
                <h2>{challenge.title}</h2>
                <p>{challenge.description}</p>
                <form onSubmit={handleCompleteChallenge}>
                  <ul>
                    {teams.map((team) => (
                      <li key={team.team_id}>
                        <h3>{team.team_name}</h3>
                        <p>{team.points} points</p>

                        {playerId !== null && playerId === hostPlayerId && (
                          <label>
                            Points for this challenge
                            <input
                              type='number'
                              min='0'
                              step='1'
                              required
                              value={challengePoints[team.team_id] ?? ''}
                              onChange={(event) =>
                                setChallengePoints((previous) => ({
                                  ...previous, [team.team_id]: event.target.value,
                                }))
                              }
                            />
                          </label>
                        )}
                      </li>
                    ))}
                  </ul>
                  {playerId !== null && playerId === hostPlayerId && (
                    <button type='submit'>
                      Complete Challenge
                    </button>
                  )}
                </form>
              </>
            ) : (
              <p>Loading challenge...</p>
            )
          ) : question ? (
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
