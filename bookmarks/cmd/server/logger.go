package main

import (
	"fmt"
	"os"

	"github.com/bihe/bookmarks/internal/config"
	log "github.com/sirupsen/logrus"
)

func setupLog(config config.AppConfig) {
	if config.Environment != "Development" {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})

		var file *os.File
		file, err := os.OpenFile(config.Logging.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", config.Logging.FilePath, err))
		}
		log.SetOutput(file)
	}
	level, err := log.ParseLevel(config.Logging.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("cannot use supplied level '%s' as a loglevel: %v", config.Logging.LogLevel, err))
	}
	log.SetLevel(level)
}
