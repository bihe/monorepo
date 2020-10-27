package filestore

import (
	"github.com/go-kit/kit/log"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
)

// ServiceMiddleware describes a service (as opposed to endpoint) middleware.
// it is used to intercept the method execution and perform actions before/after the
// serivce method execution
type ServiceMiddleware func(FileService) FileService

// ServiceLoggingMiddleware takes a logger as a dependency
// and returns a ServiceLoggingMiddleware.
func ServiceLoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next FileService) FileService {
		return loggingMiddleware{logger, next}
	}
}

// compile guard for Service implementation
var (
	_ FileService = &loggingMiddleware{}
)

type loggingMiddleware struct {
	logger log.Logger
	next   FileService
}

func (l loggingMiddleware) InitClient() (err error) {
	defer shared.Log(l.logger, "InitClient-End", err)
	return l.next.InitClient()
}

func (l loggingMiddleware) SaveFile(file FileItem) (err error) {
	shared.Log(l.logger, "SaveFile", err, "param:file", file)
	defer shared.Log(l.logger, "SaveFile-End", err)
	return l.next.SaveFile(file)
}

func (l loggingMiddleware) GetFile(filePath string) (item FileItem, err error) {
	shared.Log(l.logger, "GetFile", err, "param:filePath", filePath)
	defer shared.Log(l.logger, "GetFile-End", err)
	return l.next.GetFile(filePath)
}

func (l loggingMiddleware) DeleteFile(filePath string) (err error) {
	shared.Log(l.logger, "DeleteFile", err, "param:filePath", filePath)
	defer shared.Log(l.logger, "DeleteFile-End", err)
	return l.next.DeleteFile(filePath)
}
