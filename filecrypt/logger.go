package filecrypt

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/pkg/logging"
)

// SetupLog initiates the logging
func SetupLog(conf AppConfig) *log.Entry {
	logger := log.New()
	logger.SetOutput(os.Stdout)
	if conf.Environment != Development {
		logger.SetFormatter(&log.JSONFormatter{})

		var file *os.File
		file, err := os.OpenFile(conf.Logging.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", conf.Logging.FilePath, err))
		}
		logger.SetOutput(file)
	}
	level, err := log.ParseLevel(conf.Logging.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("cannot use supplied level '%s' as a loglevel: %v", conf.Logging.LogLevel, err))
	}
	logger.SetLevel(level)
	return logging.NewLog(logger, conf.ServiceName, conf.HostID)
}
