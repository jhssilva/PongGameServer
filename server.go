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

type Message struct {
	Message string `json:"message"`
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Create a hub
	hub := NewHub()

	// Start a go routine
	go hub.run()
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

	e.Logger.Fatal(e.Start(":8080"))
}

func read(hub *Hub, client *websocket.Conn) {
	for {
		var message Message
		err := client.ReadJSON(&message)
		if !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
			delete(hub.clients, client)
			break
		}
		log.Println(message)

		// Send a message to hub
		hub.broadcast <- message
	}
}
