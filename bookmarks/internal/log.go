package internal

import log "github.com/sirupsen/logrus"

const LogFieldNameFunction = "function"

func LogFunction(name string) *log.Entry {
	return log.WithField(LogFieldNameFunction, name)
}
