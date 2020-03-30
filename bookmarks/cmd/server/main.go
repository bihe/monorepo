package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/bihe/bookmarks/internal"
	"github.com/bihe/bookmarks/internal/config"
	"github.com/bihe/bookmarks/internal/server"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wangii/emoji"

	log "github.com/sirupsen/logrus"
)

var (
	// Version exports the application version
	Version = "1.0.0"
	// Build provides information about the application build
	Build = "202001121400"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// run configures and starts the Server
func run() (err error) {
	version := internal.VersionInfo{
		Version: Version,
		Build:   Build,
	}

	hostName, port, basePath, appConfig := readConfig()
	l := setupLog(appConfig)
	apiSrv := server.Create(basePath, appConfig, version, l)
	addr := fmt.Sprintf("%s:%d", hostName, port)
	httpSrv := &http.Server{Addr: addr, Handler: apiSrv}

	go func() {
		fmt.Printf("%s Starting server ...\n", emoji.EmojiTagToUnicode(`:rocket:`))
		fmt.Printf("%s Version: '%s-%s'\n", emoji.EmojiTagToUnicode(`:bookmark:`), Version, Build)
		fmt.Printf("%s Environment: '%s'\n", emoji.EmojiTagToUnicode(`:white_check_mark:`), appConfig.Environment)
		fmt.Printf("%s Listening on '%s'\n", emoji.EmojiTagToUnicode(`:computer:`), httpSrv.Addr)
		fmt.Printf("%s Ready!\n", emoji.EmojiTagToUnicode(`:checkered_flag:`))

		if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			return
		}
	}()
	return graceful(httpSrv, 5*time.Second)
}

func graceful(s *http.Server, timeout time.Duration) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.Infof("\nShutdown with timeout: %s\n", timeout)
	if err := s.Shutdown(ctx); err != nil {
		return err
	}

	log.Info("Server stopped")
	return nil
}

// --------------------------------------------------------------------------
// internal logic / helpers
// --------------------------------------------------------------------------

func readConfig() (hostname string, port int, basePath string, conf config.AppConfig) {
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
	viper.SetEnvPrefix("bm")                            // use this prefix for environment variabls to overwrite
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Could not get server configuration values: %v", err))
	}
	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Sprintf("Could not unmarshall server configuration values: %v", err))
	}
	return
}
