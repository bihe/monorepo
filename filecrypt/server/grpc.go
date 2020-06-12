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

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.binggl.net/monorepo/filecrypt"
	"golang.binggl.net/monorepo/pkg/server"
	"golang.binggl.net/monorepo/proto"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

// Run is the entry-point for the filecrypt service
// where initialization, setup and execution is done
func Run(version, build string) error {
	hostname, port, basePath, appConfig := readConfig()
	addr := fmt.Sprintf("%s:%d", hostname, port)
	l := filecrypt.SetupLog(appConfig)

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	svc := NewFileCrypter(basePath, l, appConfig.TokenSecurity)
	proto.RegisterCrypterServer(srv, svc)

	go func() {
		server.PrintServerBanner("filecrypt", version, build, string(appConfig.Environment), addr)
		if err := srv.Serve(conn); err != http.ErrServerClosed {
			return
		}
	}()
	return graceful(srv, l, 5*time.Second)
}

func graceful(s *grpc.Server, logger *log.Entry, timeout time.Duration) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Infof("\nShutdown with timeout: %s\n", timeout)
	s.GracefulStop()
	logger.Info("Server stopped")
	return nil
}

// --------------------------------------------------------------------------

func readConfig() (hostname string, port int, basePath string, conf filecrypt.AppConfig) {
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
