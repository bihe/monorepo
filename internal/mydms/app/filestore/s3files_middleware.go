package filestore

import (
	"golang.binggl.net/monorepo/pkg/logging"
)

// ServiceMiddleware describes a service (as opposed to endpoint) middleware.
// it is used to intercept the method execution and perform actions before/after the
// serivce method execution
type ServiceMiddleware func(FileService) FileService

// ServiceLoggingMiddleware takes a logger as a dependency
// and returns a ServiceLoggingMiddleware.
func ServiceLoggingMiddleware(logger logging.Logger) ServiceMiddleware {
	return func(next FileService) FileService {
		return loggingMiddleware{logger, next}
	}
}

// compile guard for Service implementation
var (
	_ FileService = &loggingMiddleware{}
)

type loggingMiddleware struct {
	logger logging.Logger
	next   FileService
}

func (l loggingMiddleware) InitClient() (err error) {
	defer l.logger.Info("called InitClient", logging.ErrV(err))
	return l.next.InitClient()
}

func (l loggingMiddleware) SaveFile(file FileItem) (err error) {
	l.logger.Info("SaveFile", logging.LogV("param:file", file.String()))
	defer l.logger.Info("called SaveFile", logging.ErrV(err))
	return l.next.SaveFile(file)
}

func (l loggingMiddleware) GetFile(filePath string) (item FileItem, err error) {
	l.logger.Info("GetFile", logging.LogV("param:filePath", filePath))
	defer l.logger.Info("called GetFile", logging.ErrV(err))
	return l.next.GetFile(filePath)
}

func (l loggingMiddleware) DeleteFile(filePath string) (err error) {
	l.logger.Info("DeleteFile", logging.LogV("param:filePath", filePath))
	defer l.logger.Info("called DeleteFile", logging.ErrV(err))
	return l.next.DeleteFile(filePath)
}
