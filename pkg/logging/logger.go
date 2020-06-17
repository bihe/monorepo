package logging

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// LogConfig specifies logging settings for
type LogConfig struct {
	FilePath string
	LogLevel string
	Trace    TraceConfig
}

// TraceConfig is used to correlate logging-entries
type TraceConfig struct {
	AppName string
	HostID  string
}

// Setup usesd the supplied configuration to setup application-logging
func Setup(conf LogConfig, env string) *log.Entry {
	logger := log.New()
	if strings.ToLower(env) != "development" {
		logger.SetFormatter(&log.JSONFormatter{})

		var file *os.File
		file, err := os.OpenFile(conf.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", conf.FilePath, err))
		}
		logger.SetOutput(file)
	}
	level, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("cannot use supplied level '%s' as a loglevel: %v", conf.LogLevel, err))
	}
	logger.SetLevel(level)
	return NewLog(logger, conf.Trace.AppName, conf.Trace.HostID)
}
