package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type PlayersGameData struct {
	Key string `json:"key"`
}
type Dimensions struct {
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
}

type Coord struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type Object struct {
	Size     Dimensions `json:"size"`
	Position Coord      `json:"pos"`
	Speed    Coord
	Mass     float32
}

type Particle struct {
	Position Vector2    `json:"pos"`
	Size     Dimensions `json:"size"`
	Velocity Vector2
	Mass     float32
	Dt       float32
}

type Board struct {
	Size Dimensions `json:"size"`
	Bar  Dimensions `json:"bar"`
}

type Player struct {
	Bar   Object `json:"bar"`
	Score int    `json:"score"`
}

type ServerGameData struct {
	GameStatus bool     `json:"gameStatus"`
	Board      Board    `json:"board"`
	Ball       Particle `json:"ball"`
	Player1    Player   `json:"player1"`
	Player2    Player   `json:"player2"`
}

// Vectors
type Vector2 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// Create a echo
var e = echo.New()

// Create a hub
var hub = NewHub()

// Game Data
var gameData ServerGameData

func setConfigurationEcho() {
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	setRoots()
}

func setRoots() {
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Pong Game!")
	})

	e.GET("/join", func(c echo.Context) error {
		var nrClients = len(hub.clients)
		log.Println(nrClients)
		if (nrClients) > 2 {
			return c.String(http.StatusForbidden, "You can't join the game. There is already 2 players!")
		} else if nrClients == 2 {
			//go startGame()
		}

		// Test Code
		go startGame()
		// Test Code

		return c.String(http.StatusOK, "Join the Game Successfuly!")

	})

	e.GET("/ws", func(c echo.Context) error {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if !errors.Is(err, nil) {
			log.Println(err)
		}

		defer func() {
			delete(hub.clients, ws)
			gameData.GameStatus = false
			ws.Close()
			log.Printf("Closed!")
		}()

		// Add client
		hub.clients[ws] = true

		log.Println("Connected!")

		// Listen on connection
		read(hub, ws)
		return nil
	})
}

func handlerPlayerKeyPress(playerData PlayersGameData) {
	//var message ServerGameData
	var key = playerData.Key
	switch key {
	case "w":
		if gameData.Player1.Bar.Position.Y > 0 {
			gameData.Player1.Bar.Position.Y -= gameData.Player1.Bar.Speed.Y
		} else {
			return
		}
	case "s":
		if gameData.Player1.Bar.Position.Y+gameData.Player1.Bar.Size.Height < gameData.Board.Size.Height {
			gameData.Player1.Bar.Position.Y += gameData.Player1.Bar.Speed.Y
		} else {
			return
		}
	case "ArrowUp":
		if gameData.Player2.Bar.Position.Y > 0 {
			gameData.Player2.Bar.Position.Y -= gameData.Player2.Bar.Speed.Y
		} else {
			return
		}
	case "ArrowDown":
		if gameData.Player2.Bar.Position.Y+gameData.Player2.Bar.Size.Height < gameData.Board.Size.Height {
			gameData.Player2.Bar.Position.Y += gameData.Player2.Bar.Speed.Y
		} else {
			return
		}
	default:
		log.Println("The key sent can't be handled by the server side. Accepted Keys are -> 'w', 's', 'ArrowDown', 'ArrowUp' ")
		return
	}
	//sendGameDataMessage()
}

func read(hub *Hub, client *websocket.Conn) {
	for {
		var playersGameDataMessages PlayersGameData
		err := client.ReadJSON(&playersGameDataMessages)
		if !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
			delete(hub.clients, client)
			break
		}

		handlerPlayerKeyPress(playersGameDataMessages)
	}
}

func game() {
	gameData.GameStatus = true
	log.Println("Game Running")
	gameLoop()
	log.Println("Game Ended")
}

func increasePointsPlayer(player int) {
	if player == 1 {
		gameData.Player1.Score += 1
	} else {
		gameData.Player2.Score += 1
	}

}

func isBallCollidedPlayers(player *int) bool {
	var ball = gameData.Ball
	var player1 = gameData.Player1
	var player2 = gameData.Player2

	// Check colision with Player 1
	if ball.Position.X <= (player1.Bar.Position.X+player1.Bar.Size.Width) && ball.Position.X >= (player1.Bar.Position.X) {
		if ball.Position.Y <= player1.Bar.Position.Y+player1.Bar.Size.Height && ball.Position.Y+ball.Size.Height >= player1.Bar.Position.Y {
			*player = 1
			return true
		}
	}

	// Check colision with Player 2
	if ball.Position.X+ball.Size.Width >= player2.Bar.Position.X && ball.Position.X <= player2.Bar.Position.X+player2.Bar.Size.Width {
		if ball.Position.Y <= player2.Bar.Position.Y+player1.Bar.Size.Height && ball.Position.Y+ball.Size.Height >= player2.Bar.Position.Y {
			*player = 2
			return true
		}
	}

	return false
}

func handlerBallColisionWithXAxis() {
	// Detect colision with X Limits
	var ball = &gameData.Ball
	var board_limit = &gameData.Board.Size

	if ((*ball).Position.X) < 0 {
		increasePointsPlayer(2)
		resetBall()
		return
	} else if (*ball).Position.X+(*ball).Size.Width > (*board_limit).Width {
		increasePointsPlayer(1)
		resetBall()
		return
	}
}

func handlerBallColisionWithYAxis() {
	//Detect colision in the Y limits
	var ball = &gameData.Ball
	var board_limit = &gameData.Board.Size
	if (*ball).Position.Y < 0 || (*ball).Position.Y+(*ball).Size.Height > (*board_limit).Height {
		((*ball).Velocity.Y) *= -1
	}
}

func handlerBallColisionWithPlayers() {
	// Detect colision with the Players Bar
	var player_nr int = 0
	if isBallCollidedPlayers(&player_nr) {
		if player_nr == 1 {
			if gameData.Ball.Velocity.X < 0 {
				gameData.Ball.Velocity.X *= -1
			}
		} else if player_nr == 2 {
			if gameData.Ball.Velocity.X > 0 {
				gameData.Ball.Velocity.X *= -1
			}
		}

	}
}

func ballMovement() {
	//  Particle *particle = &particles[i];
	//             Vector2 force = ComputeForce(particle);
	//             Vector2 acceleration = (Vector2){force.x / particle->mass, force.y / particle->mass};
	//             particle->velocity.x += acceleration.x * dt;
	//             particle->velocity.y += acceleration.y * dt;
	//             particle->position.x += particle->velocity.x * dt;
	//             particle->position.y += particle->velocity.y * dt;

	particle := &gameData.Ball
	//force := ComputeForce(particle)
	//acceleration := Vector2{X: force.X / (*particle).Mass, Y: force.Y / (*particle).Mass}

	// ((*particle).Velocity.X) += acceleration.X * (*particle).Dt
	// ((*particle).Velocity.Y) += acceleration.Y * (*particle).Dt
	((*particle).Position.X) += (*particle).Velocity.X * (*particle).Dt
	((*particle).Position.Y) += (*particle).Velocity.Y * (*particle).Dt

	handlerBallColisionWithXAxis()
	handlerBallColisionWithYAxis()
	handlerBallColisionWithPlayers()

	// Ball Acceleration max 60
	if (*particle).Dt < 60 {
		(*particle).Dt += 1
	}

}

func ComputeForce(particle *Particle) Vector2 {
	return Vector2{X: 0, Y: 1} //(*particle).Mass * 9.81}
}

func resetBall() {
	gameData.Ball.Mass = 1 // 1 Kg
	gameData.Ball.Position = Vector2{X: gameData.Board.Size.Width / 2, Y: gameData.Board.Size.Height / 2}
	gameData.Ball.Velocity = Vector2{X: 0.15, Y: 0.15}
	gameData.Ball.Dt = 0
	gameData.Ball.Size.Height = 15
	gameData.Ball.Size.Width = 15
}

func gameInteraction() {
	ballMovement()
}

func gameStatus() (status bool) {
	return gameData.GameStatus
}

func gameLoop() {
	tick := time.Tick(time.Second / 60)

	for gameStatus() {
		select {
		case <-tick:
			gameInteraction()
			sendGameDataMessage()
			if !gameStatus() {
				break
			}
		}
	}
}

func sendGameDataMessage() {
	// Send a message to hub
	hub.broadcast <- gameData
}

func resetPositions() {
	// Board
	gameData.GameStatus = false
	gameData.Board.Size.Width = 650
	gameData.Board.Size.Height = 480
	gameData.Board.Bar.Width = 10
	gameData.Board.Bar.Height = 100

	// Ball
	resetBall()

	// Players
	// Player 1
	gameData.Player1.Bar.Size.Width = gameData.Board.Bar.Width
	gameData.Player1.Bar.Size.Height = gameData.Board.Bar.Height
	gameData.Player1.Bar.Position.X = (gameData.Player1.Bar.Size.Width / 2)
	gameData.Player1.Bar.Position.Y = (gameData.Board.Size.Height / 2) - (gameData.Player1.Bar.Size.Height / 2)
	gameData.Player1.Bar.Speed.Y = 30.0
	gameData.Player1.Bar.Speed.X = 0
	gameData.Player1.Score = 0

	// Player 2
	gameData.Player2.Bar.Size.Width = gameData.Board.Bar.Width
	gameData.Player2.Bar.Size.Height = gameData.Board.Bar.Height
	gameData.Player2.Bar.Position.X = gameData.Board.Size.Width - (1.5 * gameData.Player2.Bar.Size.Width)
	gameData.Player2.Bar.Position.Y = (gameData.Board.Size.Height / 2) - (gameData.Player2.Bar.Size.Height / 2)
	gameData.Player2.Bar.Speed.Y = 10.0
	gameData.Player2.Bar.Speed.X = 0
	gameData.Player1.Score = 2
}

func startGame() {
	resetPositions()
	go game()
}

func main() {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Start a go routine
	go hub.run()
	setConfigurationEcho()

	e.Logger.Fatal(e.Start(":8080"))
}
