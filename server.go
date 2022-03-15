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
	Radius float32 `json:"radius"`
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
			return c.String(http.StatusForbidden, "There is already 2 players!")
		} else if nrClients == 2 {
			//go startGame()
		}

		// Test Code
		go startGame()
		// Test Code

		return c.String(http.StatusOK, "Player Joined the game!")

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

func handlePlayerMovement(playerData PlayersGameData) {
	//var message ServerGameData
	var key = playerData.Key
	switch key {
	case "w":
		gameData.Player1.Bar.Pos.Y -= gameData.Player1.Bar.Speed.Y
	case "s":
		gameData.Player1.Bar.Pos.Y += gameData.Player1.Bar.Speed.Y
	case "ArrowUp":
		gameData.Player2.Bar.Pos.Y -= gameData.Player2.Bar.Speed.Y
	case "ArrowDown":
		gameData.Player2.Bar.Pos.Y += gameData.Player2.Bar.Speed.Y
	default:
		log.Println("The key sent can't be handled by the server side. Accepted Keys are -> 'w', 's', 'ArrowDown', 'ArrowUp' ")
		return
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
		//log.Println(playersGameDataMessages)
		handlePlayerMovement(playersGameDataMessages)
	}
}

func game() {
	gameData.GameStatus = true
	log.Println("Game Running")
	gameLoop()
	log.Println("Game Ended")
}

func gameInteraction() {
	gameData.Ball.Pos.X += gameData.Ball.Speed.X
	gameData.Ball.Pos.Y += gameData.Ball.Speed.Y

	if gameData.Ball.Pos.X > gameData.Board.Size.Width || gameData.Ball.Pos.X < 0 {
		gameData.Ball.Speed.X *= -1
	} else if gameData.Ball.Pos.Y < 0 || gameData.Ball.Pos.Y > gameData.Board.Size.Height {
		gameData.Ball.Speed.Y *= -1
	}

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
	gameData.GameStatus = false
	gameData.Board.Size.Width = 650
	gameData.Board.Size.Height = 480
	gameData.Board.Bar.Width = 20
	gameData.Board.Bar.Height = 100

	gameData.Ball.Pos.X = gameData.Board.Size.Width / 2
	gameData.Ball.Pos.Y = gameData.Board.Size.Height / 2
	gameData.Ball.Speed.X = 5
	gameData.Ball.Speed.Y = 5
	gameData.Ball.Size.Radius = 10

	gameData.Player1.Bar.Size.Width = gameData.Board.Bar.Width
	gameData.Player1.Bar.Size.Height = gameData.Board.Bar.Height

	gameData.Player1.Bar.Pos.X = (gameData.Player1.Bar.Size.Width / 2)
	gameData.Player1.Bar.Pos.Y = (gameData.Board.Size.Width / 2) - (gameData.Player1.Bar.Size.Width / 2)
	gameData.Player1.Bar.Speed.Y = 40
	gameData.Player1.Bar.Speed.X = 0

	gameData.Player2.Bar.Size.Width = gameData.Board.Bar.Width
	gameData.Player2.Bar.Size.Height = gameData.Board.Bar.Height
	gameData.Player2.Bar.Pos.X = gameData.Board.Size.Width - gameData.Player2.Bar.Size.Width - 10
	gameData.Player2.Bar.Pos.Y = (gameData.Board.Size.Height / 2) - (gameData.Player2.Bar.Size.Height / 2)
	gameData.Player2.Bar.Speed.Y = 40
	gameData.Player2.Bar.Speed.X = 0

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
