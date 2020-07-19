package server

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

	"github.com/go-kit/kit/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.binggl.net/monorepo/crypter"
	"golang.binggl.net/monorepo/pkg/server"
	"golang.binggl.net/monorepo/proto"
	"google.golang.org/grpc"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

// Run is the entry-point for the cryper service
// where initialization, setup and execution is done
func Run(version, build string) error {

	hostname, port, _, config := readConfig()
	addr := fmt.Sprintf("%s:%d", hostname, port)

	// Create a single logger, which we'll use and give to other components.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

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
		logger.Log("transport", "gRPC", "during", "Listen", "err", err)
		return fmt.Errorf("could not start grpc listener: %v", err)
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	proto.RegisterCrypterServer(srv, grpcServer)

	go func() {
		server.PrintServerBanner("crypter", version, build, string(config.Environment), addr)
		if err := srv.Serve(grpcListener); err != http.ErrServerClosed {
			return
		}
	}()
	return graceful(srv, logger, 5*time.Second)
}

// --------------------------------------------------------------------------

func graceful(s *grpc.Server, logger log.Logger, timeout time.Duration) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Log("\nShutdown with timeout: %s\n", timeout)
	s.GracefulStop()
	logger.Log("Server stopped")
	return nil
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