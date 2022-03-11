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
	X string `json:"x"`
	Y string `json:"y"`
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

func setRoots() {
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/ws", func(c echo.Context) error {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if !errors.Is(err, nil) {
			log.Println(err)
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

func main() {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Start a go routine
	go hub.run()
	setRoots()

	e.Logger.Fatal(e.Start(":8080"))
}

func read(hub *Hub, client *websocket.Conn) {
	for {
		var message ServerGameData
		var playersGameDataMessages PlayersGameData
		err := client.ReadJSON(&playersGameDataMessages)
		if !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
			delete(hub.clients, client)
			break
		}
		log.Println(playersGameDataMessages)

		message.Ball.X = "10"
		message.Ball.Y = "20"

		// Send a message to hub
		hub.broadcast <- message
	}
}
