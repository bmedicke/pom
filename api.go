package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func runServer(config Config, command chan pomodoroCommand) {
	server := echo.New()
	// silence the server:
	server.HideBanner = true
	server.Logger.SetLevel(log.OFF)

	server.GET("/continue", func(c echo.Context) error {
		command <- pomodoroCommand{commandtype: "continue"}
		return c.String(http.StatusOK, `{"status":"command_sent"}`)
	})
	server.Start(config.Server)
}
