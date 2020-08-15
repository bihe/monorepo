package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

// LogConfig specifies logging settings for
type LogConfig struct {
	FilePath string
	LogLevel string
	Trace    TraceConfig
	// GrayLogServer defines the address of a log-aggregator using Graylog
	GrayLogServer string
}

// TraceConfig is used to correlate logging-entries
type TraceConfig struct {
	AppName string
	HostID  string
}

// Setup usesd the supplied configuration to setup application-logging
func Setup(conf LogConfig, env string) *log.Entry {
	var (
		file       *os.File
		writer     io.Writer
		gelfWriter gelf.Writer
		err        error
	)

	logger := log.New()
	if strings.ToLower(env) != "development" {
		logger.SetFormatter(&log.JSONFormatter{})
		file, err = os.OpenFile(conf.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", conf.FilePath, err))
		}
		writer = file
		if conf.GrayLogServer != "" {
			gelfWriter, err = gelf.NewUDPWriter(conf.GrayLogServer)
			if err != nil {
				panic(fmt.Sprintf("could not create a new gelf UDP writer: %s", err))
			}
			// log to both file and graylog2
			writer = io.MultiWriter(file, gelfWriter)
		}
		logger.SetOutput(writer)
	}
	level, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("cannot use supplied level '%s' as a loglevel: %v", conf.LogLevel, err))
	}
	logger.SetLevel(level)
	return NewLog(logger, conf.Trace.AppName, conf.Trace.HostID)
}
