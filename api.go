package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func runServer(config Config, command chan pomodoroCommand, pom *pomodoro) {
	server := echo.New()
	// allow cross origin requests:
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// silence the server:
	server.HideBanner = true
	server.Logger.SetLevel(log.OFF)

	server.Static("/live", "static/")

	server.GET("/continue", func(c echo.Context) error {
		command <- pomodoroCommand{commandtype: "continue"}
		return c.String(http.StatusOK, `{"status":"command_sent"}`)
	})

	server.GET("/ws", func(c echo.Context) error {
		tick := time.Tick(time.Millisecond * 500)
		ws, _ := upgrader.Upgrade(c.Response(), c.Request(), nil)
		defer ws.Close()

		for {
			<-tick
			pomJSON, _ := json.MarshalIndent(*pom, "", "  ")
			ws.WriteMessage(websocket.TextMessage, []byte(pomJSON))
		}
	})

	server.Start(config.Server)
}
