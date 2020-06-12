package server

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.binggl.net/monorepo/mydms/config"
	"golang.binggl.net/monorepo/pkg/logging"

	chi "github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/pkg/handler"
)

func setupLog(conf config.AppConfig, e *echo.Echo) *log.Entry {
	logger := log.New()
	logEntry := logging.NewLog(logger, conf.AppName, conf.HostID)

	switch conf.Environment {
	case config.Development:
		logger.SetOutput(os.Stdout)
		logger.SetLevel(log.DebugLevel)
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "[${time_rfc3339_nano}] (${id}) ${method} '${uri}' [${status}] Host: ${host}, IP: ${remote_ip}, error: '${error}', (latency: ${latency_human}) \n",
		}))
	case config.Production:
		logger.SetFormatter(&log.JSONFormatter{})
		var file *os.File
		file, err := os.OpenFile(conf.Logging.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", conf.Logging.FilePath, err))
		}
		logger.SetOutput(file)

		level, err := log.ParseLevel(conf.Logging.LogLevel)
		if err != nil {
			panic(fmt.Sprintf("cannot use supplied level '%s' as a loglevel: %v", conf.Logging.LogLevel, err))
		}
		logger.SetLevel(level)

		switch level {
		case log.TraceLevel:
			fallthrough
		case log.DebugLevel:
			fallthrough
		case log.InfoLevel:
			e.Use(echo.WrapMiddleware(chi.RequestID)) // my logging expects the go-chi request-ID middleware!
			e.Use(echo.WrapMiddleware(handler.NewLoggerMiddleware(logEntry).LoggerContext))
		}
	}
	return logEntry
}
