package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

	home, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	server.Static("/live", filepath.Join(home, ".config/pom/static"))

	server.GET("/state", func(c echo.Context) error {
		return c.String(http.StatusOK, getStatusJSON(pom))
	})

	server.POST("/continue", func(c echo.Context) error {
		command <- pomodoroCommand{commandtype: "continue"}
		return c.String(http.StatusOK, getStatusJSON(pom))
	})

	server.GET("/continue", func(c echo.Context) error {
		command <- pomodoroCommand{commandtype: "continue"}
		return c.String(http.StatusOK, getStatusJSON(pom))
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

func getStatusJSON(pom *pomodoro) string {
	return `{"active":` + strconv.FormatBool(!(*pom).waiting) + `}`
}
