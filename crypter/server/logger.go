package server

import (
	"fmt"
	"io"
	"os"

	"github.com/go-kit/kit/log"
	"golang.binggl.net/monorepo/crypter"
)

// SetupLog initiates the logger
func SetupLog(config crypter.AppConfig) (log.Logger, io.WriteCloser) {
	var (
		file   *os.File
		logger log.Logger
	)
	logger = log.NewLogfmtLogger(os.Stderr)
	if config.Environment != crypter.Development {
		file, err := os.OpenFile(config.Logging.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", config.Logging.FilePath, err))
		}
		logger = log.NewLogfmtLogger(file)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	return logger, file
}
