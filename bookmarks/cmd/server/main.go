package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bihe/bookmarks/internal"
	"github.com/bihe/bookmarks/internal/config"
	"github.com/bihe/bookmarks/internal/server"
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

	args := parseFlags()
	appConfig := configFromFile(args.ConfigFile)
	apiSrv := server.Create(args.BasePath, appConfig, version, args.Environment)

	if args.Environment != "" {
		appConfig.Environment = args.Environment
	}

	setupLog(appConfig)
	addr := fmt.Sprintf("%s:%d", args.HostName, args.Port)
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

// args is used to configure the API server
type args struct {
	HostName    string
	Port        int
	ConfigFile  string
	BasePath    string
	Environment string
}

func parseFlags() *args {
	c := new(args)
	flag.StringVar(&c.HostName, "hostname", "localhost", "the server hostname")
	flag.IntVar(&c.Port, "port", 3000, "network port to listen")
	flag.StringVar(&c.BasePath, "b", "./", "the base path of the application")
	flag.StringVar(&c.ConfigFile, "c", "application.json", "path to the application c file")
	flag.StringVar(&c.Environment, "e", "Development", "name of the environment to use")
	flag.Parse()
	return c
}

func configFromFile(configFileName string) config.AppConfig {
	if !fileExists(configFileName) {
		// if the given filename does not exists, use the filename from an environment variable
		// if that fails as well, the logic will panic below
		configFileName = os.Getenv("CONFIG_FILE_NAME")
	}
	f, err := os.Open(configFileName)
	if err != nil {
		panic(fmt.Sprintf("Could not open specific config file '%s': %v", configFileName, err))
	}
	defer f.Close()

	c, err := config.GetSettings(f)
	if err != nil {
		panic(fmt.Sprintf("Could not get server config values from file '%s': %v", configFileName, err))
	}
	return *c
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
