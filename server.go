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

type Speed struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type Bar struct {
	Size  Dimensions `json:"size"`
	Speed Speed      `json:"speed"`
	Pos   Coord      `json:"pos"`
}

type Ball struct {
	Size  Dimensions `json:"size"`
	Pos   Coord      `json:"pos"`
	Speed Speed      `json:"speed"`
}

type Board struct {
	Size Dimensions `json:"size"`
	Bar  Dimensions `json:"bar"`
}

type Player struct {
	Bar   Bar `json:"bar"`
	Score int `json:"score"`
}

type ServerGameData struct {
	GameStatus bool   `json:"gameStatus"`
	Board      Board  `json:"board"`
	Ball       Ball   `json:"ball"`
	Player1    Player `json:"player1"`
	Player2    Player `json:"player2"`
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

func handlePlayerKeyPress(playerData PlayersGameData) {
	//var message ServerGameData
	var key = playerData.Key
	switch key {
	case "w":
		if gameData.Player1.Bar.Pos.Y > 0 {
			gameData.Player1.Bar.Pos.Y -= gameData.Player1.Bar.Speed.Y
		} else {
			return
		}
	case "s":
		if gameData.Player1.Bar.Pos.Y+gameData.Player1.Bar.Size.Height < gameData.Board.Size.Height {
			gameData.Player1.Bar.Pos.Y += gameData.Player1.Bar.Speed.Y
		} else {
			return
		}
	case "ArrowUp":
		if gameData.Player2.Bar.Pos.Y > 0 {
			gameData.Player2.Bar.Pos.Y -= gameData.Player2.Bar.Speed.Y
		} else {
			return
		}
	case "ArrowDown":
		if gameData.Player2.Bar.Pos.Y+gameData.Player2.Bar.Size.Height < gameData.Board.Size.Height {
			gameData.Player2.Bar.Pos.Y += gameData.Player2.Bar.Speed.Y
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

		handlePlayerKeyPress(playersGameDataMessages)
	}
}

func game() {
	gameData.GameStatus = true
	log.Println("Game Running")
	gameLoop()
	log.Println("Game Ended")
}

func increasePointsPlayer1() {
	gameData.Player1.Score += 1
}

func increasePointsPlayer2() {
	gameData.Player2.Score += 1
}

func resetBall() {
	gameData.Ball.Pos.X = gameData.Board.Size.Width / 2
	gameData.Ball.Pos.Y = gameData.Board.Size.Height / 2
	gameData.Ball.Speed.X = 5
	gameData.Ball.Speed.Y = 5
}

func isBallCollidedPlayers(player *int) bool {
	var ball = gameData.Ball
	var player1 = gameData.Player1
	var player2 = gameData.Player2

	// Check colision with Player 1
	if ball.Pos.X <= (player1.Bar.Pos.X+player1.Bar.Size.Width) && ball.Pos.X >= (player1.Bar.Pos.X) {
		if ball.Pos.Y <= player1.Bar.Pos.Y+player1.Bar.Size.Height && ball.Pos.Y+ball.Size.Height >= player1.Bar.Pos.Y {
			*player = 1
			return true
		}
	}

	// Check colision with Player 2
	if ball.Pos.X+ball.Size.Width >= player2.Bar.Pos.X && ball.Pos.X <= player2.Bar.Pos.X+player2.Bar.Size.Width {
		if ball.Pos.Y <= player2.Bar.Pos.Y+player1.Bar.Size.Height && ball.Pos.Y+ball.Size.Height >= player2.Bar.Pos.Y {
			*player = 2
			return true
		}
	}

	return false
}

func ballInteraction() {
	var pos_x = &gameData.Ball.Pos.X
	var pos_y = &gameData.Ball.Pos.Y
	var ball_size_x = gameData.Ball.Size.Width
	var ball_size_y = gameData.Ball.Size.Height
	var max_x = gameData.Board.Size.Width
	var max_y = gameData.Board.Size.Height
	var player_nr = 0

	//Detect colision in the Y limits
	if (*pos_y) < 0 || (*pos_y)+ball_size_y > max_y {
		gameData.Ball.Speed.Y *= -1
	}

	// Detect colision with the Players Bar
	if isBallCollidedPlayers(&player_nr) {
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

	// Ball Movement
	(*pos_x) += gameData.Ball.Speed.X
	(*pos_y) += gameData.Ball.Speed.Y

	// Detect colision with X Limits
	if (*pos_x) < 0 {
		increasePointsPlayer2()
		resetBall()
		return
	} else if (*pos_x)+ball_size_x > max_x {
		increasePointsPlayer1()
		resetBall()
		return
	}
}

func gameInteraction() {
	ballInteraction()
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
	gameData.Ball.Pos.X = gameData.Board.Size.Width / 2
	gameData.Ball.Pos.Y = gameData.Board.Size.Height / 2
	gameData.Ball.Speed.X = 5
	gameData.Ball.Speed.Y = 5
	gameData.Ball.Size.Height = 15
	gameData.Ball.Size.Width = 15

	// Players
	// Player 1
	gameData.Player1.Bar.Size.Width = gameData.Board.Bar.Width
	gameData.Player1.Bar.Size.Height = gameData.Board.Bar.Height
	gameData.Player1.Bar.Pos.X = (gameData.Player1.Bar.Size.Width / 2)
	gameData.Player1.Bar.Pos.Y = (gameData.Board.Size.Height / 2) - (gameData.Player1.Bar.Size.Height / 2)
	gameData.Player1.Bar.Speed.Y = 10.0
	gameData.Player1.Bar.Speed.X = 0
	gameData.Player1.Score = 0

	// Player 2
	gameData.Player2.Bar.Size.Width = gameData.Board.Bar.Width
	gameData.Player2.Bar.Size.Height = gameData.Board.Bar.Height
	gameData.Player2.Bar.Pos.X = gameData.Board.Size.Width - (1.5 * gameData.Player2.Bar.Size.Width)
	gameData.Player2.Bar.Pos.Y = (gameData.Board.Size.Height / 2) - (gameData.Player2.Bar.Size.Height / 2)
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
