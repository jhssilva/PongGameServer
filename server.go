package main

import (
	"errors"
	"log"
	"net/http"

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

type Coord struct {
	X     float32 `json:"x"`
	Y     float32 `json:"y"`
	Speed float32 `json:"speed"`
}

type ServerGameData struct {
	Ball    Coord `json:"ball"`
	Player1 Coord `json:"player1"`
	Player2 Coord `json:"player2"`
}

// Create a echo
var e = echo.New()

// Create a hub
var hub = NewHub()

func setConfigurationEcho() {
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	setRoots()
}

func setRoots() {
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/join", func(c echo.Context) error {
		if len(hub.clients)+1 > 2 {

			return c.String(http.StatusForbidden, "There is already 2 players!")
		} else if len(hub.clients) == 2 {
			go startGame()
		}
		return c.String(http.StatusOK, "Player Joined the game!")
	})

	e.GET("/ws", func(c echo.Context) error {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if !errors.Is(err, nil) {
			log.Println(err)
		}

		if len(hub.clients) > 1 {
			log.Println("Queue full! 2 Players Max!")
			return nil
		}

		defer func() {
			delete(hub.clients, ws)
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

func handlePlayerMovement(keyPressed PlayersGameData) {
	//var message ServerGameData

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
		log.Println(playersGameDataMessages)

		handlePlayerMovement(playersGameDataMessages)

		// message.Ball.X = 10.2
		// message.Ball.Y = 20.3
		// message.Player1.X = 0
		// message.Player1.X = 0

		// // Send a message to hub
		// hub.broadcast <- message
	}
}

func startGame() {
	var message ServerGameData
	// Send a message to hub
	hub.broadcast <- message
}

func main() {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Start a go routine
	go hub.run()
	setConfigurationEcho()

	e.Logger.Fatal(e.Start(":8080"))
}
