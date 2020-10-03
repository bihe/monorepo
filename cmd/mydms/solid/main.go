package main

import (
	"fmt"
	"io"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/internal/mydms"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"

	logR "github.com/go-kit/kit/log/logrus"
)

var (
	// Version exports the application version
	Version = "3.0.0"
	// Build provides information about the application build
	Build = "localbuild"
	// AppName specifies the application itself
	AppName = "mydms-solid"
	// ApplicationNameKey identifies the application in structured logging
	ApplicationNameKey = "appName"
	// HostIDKey identifies the host in structured logging
	HostIDKey = "hostID"
)

func main() {
	if err := run(Version, Build); err != nil {
		fmt.Fprintf(os.Stderr, "<< ERROR-RESULT >> '%s'\n", err)
		os.Exit(1)
	}
}

// run is the entry-point for the mydms service
// where initialization, setup and execution is done
func run(version, build string) error {
	//hostname, port, _, config := readConfig()
	hostname, port, basePath, conf := server.ReadConfig("my", func() interface{} {
		return &mydms.AppConfig{} // use the correct object to deserialize the configuration
	})
	var appCfg = conf.(*mydms.AppConfig)
	lLogger, gLogger, logFile, gelfWriter := setupLog(*appCfg)

	// ensure closing of logfile on exit
	defer func(file io.WriteCloser, gw gelf.Writer) {
		if file != nil {
			file.Close()
		}
		if gw != nil {
			gw.Close()
		}
	}(logFile, gelfWriter)

	// Build the layers of the service "onion" from the inside out. First, the
	// business logic service; then, the set of endpoints that wrap the service;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	var (
		service   = appinfo.NewService(gLogger, Version, Build)
		endpoints = mydms.MakeServerEndpoints(service, gLogger)
		apiSrv    = mydms.MakeHTTPHandler(endpoints, gLogger, lLogger, mydms.HTTPHandlerOptions{
			BasePath:     basePath,
			ErrorPath:    "/error",
			AssetConfig:  appCfg.Assets,
			CookieConfig: appCfg.Cookies,
			CorsConfig:   appCfg.Cors,
			JWTConfig:    appCfg.Security,
		})
	)

	return server.Run(server.RunOptions{
		AppName:       AppName,
		Version:       Version,
		Build:         Build,
		HostName:      hostname,
		Port:          port,
		Environment:   string(appCfg.Environment),
		ServerHandler: apiSrv,
	})
}

func setupLog(cfg mydms.AppConfig) (*logrus.Entry, log.Logger, io.WriteCloser, gelf.Writer) {
	lr, file, gelfWriter := logging.SetupLog(logging.LogConfig{
		FilePath:      cfg.Logging.FilePath,
		LogLevel:      cfg.Logging.LogLevel,
		GrayLogServer: cfg.Logging.GrayLogServer,
		Trace: logging.TraceConfig{
			AppName: cfg.AppName,
			HostID:  cfg.HostID,
		},
	}, string(cfg.Environment))

	// for now I also need the logrus instance for logic in pkg
	// in future abstract the logging, use a comman logging approache
	logger := logR.NewLogrusLogger(lr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	return lr, logger, file, gelfWriter
}
