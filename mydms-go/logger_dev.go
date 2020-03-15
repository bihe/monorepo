// +build !prod

package main

import (
	"os"

	"github.com/bihe/mydms/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

// InitLogger performs a setup for the logging mechanism
func InitLogger(conf config.LogConfig, e *echo.Echo) {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339_nano}] (${id}) ${method} '${uri}' [${status}] Host: ${host}, IP: ${remote_ip}, error: '${error}', (latency: ${latency_human}) \n",
	}))
}
