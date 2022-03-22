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
	Position     Vector2    `json:"pos"`
	Size         Dimensions `json:"size"`
	Speed        Vector2
	DefaultSpeed Vector2
	MaxSpeed     Vector2
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
		var keyDescription string = "w"
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
	if (*player).Score == 10 {
		gameData.GameStatus = false
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
		if ball.Position.Y <= player2.Bar.Position.Y+player2.Bar.Size.Height && ball.Position.Y+ball.Size.Height >= player2.Bar.Position.Y {
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
		resetBall(1)
		return
	} else if (*ball).Position.X+(*ball).Size.Width > (*board_limit).Width {
		increasePointsPlayer(&gameData.Player1)
		resetBall(2)
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
		handlerBallHasTouchedPlayer(player_nr)
		handlerBallSpeedColisionPlayers(player_nr)
	}
}

func setBallSpeedAfterColisionWithPlayer(player *Player) {
	if player == nil {
		return
	}

	ball := &gameData.Ball
	// Check if the Ball it's the Superior / Inferior / Middle Part of the Player Bar
	if (*ball).Position.Y+(*ball).Size.Height/2 < (*player).Bar.Position.Y+((*player).Bar.Size.Height/3) {
		if (*ball).Speed.Y == 0 {
			(*ball).Speed.Y = -(*ball).MaxSpeed.Y
		} else if (*ball).Speed.Y > 0 {
			(*ball).Speed.Y *= -1
		}
	} else if (*ball).Position.Y+(*ball).Size.Height/2 > (*player).Bar.Position.Y+(2*((*player).Bar.Size.Height/3)) {
		if (*ball).Speed.Y == 0 {
			(*ball).Speed.Y = (*ball).MaxSpeed.Y
		} else if (*ball).Speed.Y < 0 {
			(*ball).Speed.Y *= -1
		}
	} else {
		(*ball).Speed.Y = 0
	}

	(*ball).Speed.X *= -1
}

func handlerBallSpeedColisionPlayers(player_nr int) {
	var player *Player

	if player_nr == 1 {
		player = &gameData.Player1
	} else if player_nr == 2 {
		player = &gameData.Player2
	}

	setBallSpeedAfterColisionWithPlayer(player)
}

func handlerBallHasTouchedPlayer(player_nr int) {
	ball := &gameData.Ball
	if (*ball).HasTouchedPlayer {
		return
	}

	if player_nr == 1 {
		(*ball).Speed.X = -(*ball).MaxSpeed.X
	} else if player_nr == 2 {
		(*ball).Speed.X = (*ball).MaxSpeed.X
	}

	(*ball).HasTouchedPlayer = true
}

func ballMovement() {
	ball := &gameData.Ball
	((*ball).Position.X) += (*ball).Speed.X
	((*ball).Position.Y) += (*ball).Speed.Y

	handlerBallColisionWithXAxis()
	handlerBallColisionWithYAxis()
	handlerBallColisionWithPlayers()
}

func resetPositions() {
	// Board
	resetBoard()

	// Ball
	resetBall(0)

	// Players
	resetPlayers()
}

func resetBoard() {
	gameData.GameStatus = false
	gameData.Board.Size.Width = 640
	gameData.Board.Size.Height = 480
	gameData.Board.Bar.Width = 10
	gameData.Board.Bar.Height = 120
}

func resetBall(num int) {
	ball := &gameData.Ball
	(*ball).Position = Vector2{X: gameData.Board.Size.Width / 2, Y: gameData.Board.Size.Height / 2}
	(*ball).DefaultSpeed = Vector2{X: 2, Y: 2}
	(*ball).MaxSpeed = Vector2{X: 6, Y: 6}
	(*ball).Size.Height = 15
	(*ball).Size.Width = 15
	(*ball).HasTouchedPlayer = false
	(*ball).Speed = (*ball).DefaultSpeed

	// Ball goes to Player 1
	if num == 1 {
		(*ball).Speed.X *= -1
	}
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
