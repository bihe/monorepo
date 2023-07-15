package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.binggl.net/monorepo/internal/crypter"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
	"golang.binggl.net/monorepo/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	// Version exports the application version
	Version = "1.0.0"
	// Build provides information about the application build
	Build = "localbuild"
)

// ApplicationNameKey identifies the application in structured logging
const ApplicationNameKey = "appName"

// HostIDKey identifies the host in structured logging
const HostIDKey = "hostID"

func main() {
	if err := run(Version, Build); err != nil {
		fmt.Fprintf(os.Stderr, "<< ERROR-RESULT >> '%s'\n", err)
		os.Exit(1)
	}
}

// run is the entry-point for the cryper service
// where initialization, setup and execution is done
func run(version, build string) error {
	hostname, port, _, config := server.ReadConfig[crypter.AppConfig]("cr")
	addr := fmt.Sprintf("%s:%d", hostname, port)
	logger := logConfig(config)
	defer logger.Close()

	var (
		service       = crypter.NewService(logger, config.Security)
		crypterServer = crypter.NewServer(service, logger)
		grpcServer    = grpc.NewServer()
	)
	grpcListener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("list error", logging.ErrV(err))
		return fmt.Errorf("could not start grpc listener: %v", err)
	}
	proto.RegisterCrypterServer(grpcServer, crypterServer)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	logger.Info("start-up GRPC server")
	go func() {
		server.PrintServerBanner("crypter", version, build, string(config.Environment), addr)
		if err := grpcServer.Serve(grpcListener); err != http.ErrServerClosed {
			return
		}
	}()
	return graceful(grpcServer, logger, 5*time.Second)
}

// --------------------------------------------------------------------------

func graceful(s *grpc.Server, logger logging.Logger, timeout time.Duration) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info(fmt.Sprintf("Shutdown with timeout: %s", timeout))
	s.GracefulStop()
	logger.Info("Server stopped")
	return nil
}

func logConfig(cfg crypter.AppConfig) logging.Logger {
	return logging.New(logging.LogConfig{
		FilePath:      cfg.Logging.FilePath,
		LogLevel:      cfg.Logging.LogLevel,
		GrayLogServer: cfg.Logging.GrayLogServer,
		Trace: logging.TraceConfig{
			AppName: cfg.AppName,
			HostID:  cfg.HostID,
		},
	}, cfg.Environment)
}
