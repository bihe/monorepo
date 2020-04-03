package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/bihe/mydms/internal"
	"github.com/bihe/mydms/internal/config"
	"github.com/bihe/mydms/internal/errors"
	"github.com/bihe/mydms/internal/persistence"
	"github.com/bihe/mydms/internal/security"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	_ "github.com/bihe/mydms/docs"
	echoSwagger "github.com/swaggo/echo-swagger"

	sec "golang.binggl.net/commons/security"
)

var (
	// Version exports the application version
	Version = "2.0.0"
	// Build provides information about the application build
	Build = "20190812.164451"
)

// ServerArgs is uded to configure the API server
type ServerArgs struct {
	HostName   string
	Port       int
	ConfigFile string
}

// @title mydms API
// @version 2.0
// @description This is the API of the mydms application

// @license.name MIT License
// @license.url https://raw.githubusercontent.com/bihe/mydms-go/master/LICENSE

func main() {
	api, addr := setupAPIServer()

	// Start server
	go func() {
		fmt.Printf("starting mydms.api (%s-%s)\n", Version, Build)
		if err := api.Start(addr); err != nil {
			api.Logger.Info("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := api.Shutdown(ctx); err != nil {
		api.Logger.Fatal(err)
	}
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
	viper.SetEnvPrefix("my")                            // use this prefix for environment variabls to overwrite
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Could not get server configuration values: %v", err))
	}
	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Sprintf("Could not unmarshall server configuration values: %v", err))
	}
	return
}

func setupAPIServer() (*echo.Echo, string) {
	hostName, port, _, c := readConfig()

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = errors.CustomErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	l := setupLog(c, e)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     c.Cors.Origins,
		AllowHeaders:     c.Cors.Headers,
		AllowMethods:     c.Cors.Methods,
		AllowCredentials: c.Cors.Credentials,
		MaxAge:           c.Cors.MaxAge,
	}))

	e.Use(middleware.Secure())
	e.Use(security.JwtWithConfig(security.JwtOptions{
		JwtSecret:  c.Security.JwtSecret,
		JwtIssuer:  c.Security.JwtIssuer,
		CookieName: c.Security.CookieName,
		RequiredClaim: sec.Claim{
			Name:  c.Security.Claim.Name,
			URL:   c.Security.Claim.URL,
			Roles: c.Security.Claim.Roles,
		},
		RedirectURL:   c.Security.LoginRedirect,
		CacheDuration: c.Security.CacheDuration,
	}))

	// persistence store && application version
	con := persistence.NewConn(c.Database.ConnectionString)
	version := internal.VersionInfo{
		Version: Version,
		Build:   Build,
	}
	if err := registerRoutes(e, con, c, version, l); err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	// enable swagger for API endpoints
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e, fmt.Sprintf("%s:%d", hostName, port)
}
