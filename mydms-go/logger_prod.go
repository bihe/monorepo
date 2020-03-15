// +build prod

package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/bihe/mydms/internal/config"
)

// InitLogger performs a setup for the logging mechanism
func InitLogger(conf config.LogConfig, e *echo.Echo) {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	var file *os.File
	file, err := os.OpenFile(conf.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", conf.FilePath, err))
	}
	log.SetOutput(file)
	level, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("cannot use supplied level '%s' as a loglevel: %v", conf.FilePath, err))
	}
	log.SetLevel(level)

	switch conf.LogLevel {
	case "trace":
		fallthrough
	case "debug":
		fallthrough
	case "info":
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "[${time_rfc3339_nano}] (${id}) ${method} '${uri}' [${status}] Host: ${host}, IP: ${remote_ip}, error: '${error}', (latency: ${latency_human}) \n",
		}))
		break
	}
}
