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
		file, err := os.OpenFile(config.Log.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", config.Log.FilePath, err))
		}
		log.SetOutput(file)
	}
	level, err := log.ParseLevel(config.Log.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("cannot use supplied level '%s' as a loglevel: %v", config.Log.LogLevel, err))
	}
	log.SetLevel(level)
}
