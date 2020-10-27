package main

import (
	"fmt"
	"io"
	"os"

	stdlog "log"

	"github.com/go-kit/kit/log"
	"github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/internal/mydms"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
	"golang.binggl.net/monorepo/pkg/server"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"

	logR "github.com/go-kit/kit/log/logrus"

	// include the mysql driver for runtime
	_ "github.com/go-sql-driver/mysql"
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
	logrusLog, kitLog, logFile, gelfWriter := setupLog(*appCfg)

	// std-logging also via go-kit logging
	stdlog.SetOutput(log.NewStdlibAdapter(kitLog))

	// ensure closing of logfile on exit
	defer func(file io.WriteCloser, gw gelf.Writer) {
		if file != nil {
			file.Close()
		}
		if gw != nil {
			gw.Close()
		}
	}(logFile, gelfWriter)

	// persistence store && application version
	con := persistence.NewConnForDb("mysql", appCfg.Database.ConnectionString)
	repo, err := document.NewRepository(con)
	if err != nil {
		panic(fmt.Sprintf("cannot establish database connection: %v", err))
	}

	// Build the layers of the appInfoSvc "onion" from the inside out. First, the
	// business logic appInfoSvc; then, the set of endpoints that wrap the appInfoSvc;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	var (
		fileSvc = filestore.NewService(kitLog, filestore.S3Config{
			Bucket: appCfg.Filestore.Bucket,
			Region: appCfg.Filestore.Region,
			Key:    appCfg.Filestore.Key,
			Secret: appCfg.Filestore.Secret,
		})
		uploadClient = upload.NewClient(appCfg.Upload.EndpointURL)
		appInfoSvc   = appinfo.NewService(kitLog, Version, Build)
		docSvc       = document.NewService(kitLog, repo, fileSvc, uploadClient)
		endpoints    = mydms.MakeServerEndpoints(appInfoSvc, docSvc, fileSvc, kitLog)
		apiSrv       = mydms.MakeHTTPHandler(endpoints, kitLog, logrusLog, mydms.HTTPHandlerOptions{
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
	// in future abstract the logging, use a common logging approach
	logger := logR.NewLogrusLogger(lr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	return lr, logger, file, gelfWriter
}
