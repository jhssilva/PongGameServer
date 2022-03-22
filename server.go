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

type Object struct {
	Position Vector2    `json:"pos"`
	Size     Dimensions `json:"size"`
	Speed    Vector2
}

type Board struct {
	Size Dimensions `json:"size"`
	Bar  Dimensions `json:"bar"`
}

type Player struct {
	Bar     Object `json:"bar"`
	Score   int    `json:"score"`
	LastKey string `json:"lastKey"`
}

type Ball struct {
	Object
	HasTouchedPlayer bool `json:"hasTouchedPlayer"`
}

type ServerGameData struct {
	GameStatus bool   `json:"gameStatus"`
	Board      Board  `json:"board"`
	Ball       Ball   `json:"ball"`
	Player1    Player `json:"player1"`
	Player2    Player `json:"player2"`
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

func setLastKeyPressedInPlayer(player *Player, key string) {
	(*player).LastKey = key
}

func handlerWhenKey_W_isPressed() {
	if gameData.Player1.Bar.Position.Y > 0 {
		player := &gameData.Player1
		var keyDescription string = "arrowUp"
		handlerLastKey(player, keyDescription)
		(*player).Bar.Position.Y -= (*player).Bar.Speed.Y
	} else {
		return
	}
}

func handlerWhenKey_S_isPressed() {
	if gameData.Player1.Bar.Position.Y+gameData.Player1.Bar.Size.Height < gameData.Board.Size.Height {
		var keyDescription string = "s"
		player := &gameData.Player1
		handlerLastKey(player, keyDescription)
		(*player).Bar.Position.Y += (*player).Bar.Speed.Y
	} else {
		return
	}
}

func handlerWhenKey_ArrowUp_isPressed() {
	if gameData.Player2.Bar.Position.Y > 0 {
		var keyDescription string = "arrowUp"
		player := &gameData.Player2
		handlerLastKey(player, keyDescription)
		(*player).Bar.Position.Y -= (*player).Bar.Speed.Y
	} else {
		return
	}
}

func handlerWhenKey_ArrowDown_isPressed() {
	if gameData.Player2.Bar.Position.Y+gameData.Player2.Bar.Size.Height < gameData.Board.Size.Height {
		var keyDescription string = "arrowDown"
		player := &gameData.Player2
		handlerLastKey(player, keyDescription)
		(*player).Bar.Position.Y += (*player).Bar.Speed.Y
	} else {
		return
	}
}

func handlerPlayerKeyPress(playerData PlayersGameData) {
	var key = playerData.Key
	switch key {
	case "w":
		handlerWhenKey_W_isPressed()
	case "s":
		handlerWhenKey_S_isPressed()
	case "ArrowUp":
		handlerWhenKey_ArrowUp_isPressed()
	case "ArrowDown":
		handlerWhenKey_ArrowDown_isPressed()
	default:
		log.Println("The key sent can't be handled by the server side. Accepted Keys are -> 'w', 's', 'ArrowDown', 'ArrowUp' ")
		return
	}
	sendGameDataMessage()
}

func handlerLastKey(player *Player, keyDescription string) {
	const maxSpeed float32 = 50
	if (*player).LastKey == keyDescription {
		if (*player).Bar.Speed.Y < maxSpeed {
			(*player).Bar.Speed.Y += 5
		}
	} else {
		setLastKeyPressedInPlayer(player, keyDescription)
	}
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

func increasePointsPlayer(player *Player) {
	(*player).Score += 1
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
		increasePointsPlayer(&gameData.Player2)
		resetBall()
		return
	} else if (*ball).Position.X+(*ball).Size.Width > (*board_limit).Width {
		increasePointsPlayer(&gameData.Player1)
		resetBall()
		return
	}
}

func handlerBallColisionWithYAxis() {
	//Detect colision in the Y limits
	var ball = &gameData.Ball
	var board_limit = &gameData.Board.Size
	if (*ball).Position.Y < 0 || (*ball).Position.Y+(*ball).Size.Height > (*board_limit).Height {
		((*ball).Speed.Y) *= -1
	}
}

func handlerBallColisionWithPlayers() {
	// Detect colision with the Players Bar
	var player_nr int = 0
	if isBallCollidedPlayers(&player_nr) {
		handlerHasTouchedPlayer()
		if player_nr == 1 {
			if gameData.Ball.Speed.X < 0 {
				gameData.Ball.Speed.X *= -1
			}
		} else if player_nr == 2 {
			if gameData.Ball.Speed.X > 0 {
				gameData.Ball.Speed.X *= -1
			}
		}

	}
}

func handlerHasTouchedPlayer() {

	if gameData.Ball.HasTouchedPlayer {
		return
	}

	speed := gameData.Ball.Speed
	var maxSpeed float32 = 6
	var maxMutiple float32 = 3
	if speed.X < maxSpeed {
		gameData.Ball.Speed.X *= maxMutiple
	}

	if speed.Y < maxSpeed {
		gameData.Ball.Speed.Y *= maxMutiple
	}

	gameData.Ball.HasTouchedPlayer = true
}

func ballMovement() {
	particle := &gameData.Ball
	((*particle).Position.X) += (*particle).Speed.X
	((*particle).Position.Y) += (*particle).Speed.Y

	handlerBallColisionWithXAxis()
	handlerBallColisionWithYAxis()
	handlerBallColisionWithPlayers()
}

func resetPositions() {
	// Board
	resetBoard()

	// Ball
	resetBall()

	// Players
	resetPlayers()
}

func resetBoard() {
	gameData.GameStatus = false
	gameData.Board.Size.Width = 640
	gameData.Board.Size.Height = 480
	gameData.Board.Bar.Width = 20
	gameData.Board.Bar.Height = 100
}

func resetBall() {
	gameData.Ball.Position = Vector2{X: gameData.Board.Size.Width / 2, Y: gameData.Board.Size.Height / 2}
	gameData.Ball.Speed = Vector2{X: 2, Y: 2}
	gameData.Ball.Size.Height = 15
	gameData.Ball.Size.Width = 15
	gameData.Ball.HasTouchedPlayer = false
}

func resetPlayers() {
	// Player 1
	resetPlayer(&gameData.Player1)
	gameData.Player1.Bar.Position.X = (gameData.Player1.Bar.Size.Width / 2)
	gameData.Player1.Bar.Position.Y = (gameData.Board.Size.Height / 2) - (gameData.Player1.Bar.Size.Height / 2)

	// Player 2
	resetPlayer(&gameData.Player2)
	gameData.Player2.Bar.Position.X = gameData.Board.Size.Width - (1.5 * gameData.Player2.Bar.Size.Width)
	gameData.Player2.Bar.Position.Y = (gameData.Board.Size.Height / 2) - (gameData.Player2.Bar.Size.Height / 2)

}

func resetPlayer(player *Player) {
	(*player).Bar.Size.Width = gameData.Board.Bar.Width
	(*player).Bar.Size.Height = gameData.Board.Bar.Height
	(*player).Bar.Speed.Y = 20
	(*player).Bar.Speed.X = 0
	(*player).Score = 0
	(*player).LastKey = ""
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
