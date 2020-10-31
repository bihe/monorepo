package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.binggl.net/monorepo/internal/crypter"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
	"golang.binggl.net/monorepo/proto"
	"google.golang.org/grpc"

	c "golang.binggl.net/monorepo/pkg/config"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
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
	hostname, port, _, config := readConfig()
	addr := fmt.Sprintf("%s:%d", hostname, port)

	logger := setupLog(config)
	defer logger.Close()

	// Build the layers of the service "onion" from the inside out. First, the
	// business logic service; then, the set of endpoints that wrap the service;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	var (
		service    = crypter.NewService(logger, config.TokenSecurity)
		endpoints  = crypter.NewEndpoints(service, logger)
		grpcServer = crypter.NewGRPCServer(endpoints, logger)
	)

	// The gRPC listener mounts the Go kit gRPC server we created.
	grpcListener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("list error", logging.ErrV(err))
		return fmt.Errorf("could not start grpc listener: %v", err)
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	proto.RegisterCrypterServer(srv, grpcServer)

	logger.Info("start-up GRPC server")
	go func() {
		server.PrintServerBanner("crypter", version, build, string(config.Environment), addr)
		if err := srv.Serve(grpcListener); err != http.ErrServerClosed {
			return
		}
	}()
	return graceful(srv, logger, 5*time.Second)
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

func setupLog(cfg crypter.AppConfig) logging.Logger {
	var env c.Environment

	switch cfg.Environment {
	case crypter.Development:
		env = c.Development
	case crypter.Production:
		env = c.Production
	}

	return logging.New(logging.LogConfig{
		FilePath:      cfg.Logging.FilePath,
		LogLevel:      cfg.Logging.LogLevel,
		GrayLogServer: cfg.Logging.GrayLogServer,
		Trace: logging.TraceConfig{
			AppName: cfg.ServiceName,
			HostID:  cfg.HostID,
		},
	}, env)
}

func readConfig() (hostname string, port int, basePath string, conf crypter.AppConfig) {
	flag.String("hostname", "localhost", "the server hostname")
	flag.Int("port", 3000, "network port to listen")
	flag.String("basepath", "./", "the base path of the application")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(fmt.Sprintf("Could not bind to command line: %v", err))
	}

	basePath = viper.GetString("basepath")
	hostname = viper.GetString("hostname")
	port = viper.GetInt("port")

	viper.SetConfigName("application")                  // name of config file (without extension)
	viper.SetConfigType("yaml")                         // type of the config-file
	viper.AddConfigPath(path.Join(basePath, "./_etc/")) // path to look for the config file in
	viper.AddConfigPath(path.Join(basePath, "./etc/"))  // path to look for the config file in
	viper.AddConfigPath(path.Join(basePath, "."))       // optionally look for config in the working directory
	viper.SetEnvPrefix("cr")                            // use this prefix for environment variabls to overwrite
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Could not get server configuration values: %v", err))
	}
	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Sprintf("Could not unmarshall server configuration values: %v", err))
	}

	return
}
