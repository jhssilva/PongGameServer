package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{}

type Message struct {
	Message string `json:"message"`
}

func main() {

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/ws", func(c echo.Context) error {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if !errors.Is(err, nil) {
			log.Println(err)
		}
		defer ws.Close()

		log.Println("Connected!")

		for {
			var message Message
			err := ws.ReadJSON(&message)
			if !errors.Is(err, nil) {
				log.Printf("error occurred: %v", err)
				break
			}
			log.Println(message)

			// send message from server
			if err := ws.WriteJSON(message); !errors.Is(err, nil) {
				log.Printf("error occurred: %v", err)
			}
		}

		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))
}
