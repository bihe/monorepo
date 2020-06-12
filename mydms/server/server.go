package server

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.binggl.net/monorepo/mydms"
	"golang.binggl.net/monorepo/mydms/config"
	"golang.binggl.net/monorepo/mydms/errors"
	"golang.binggl.net/monorepo/mydms/persistence"
	"golang.binggl.net/monorepo/mydms/security"

	echoSwagger "github.com/swaggo/echo-swagger"
	// swagger documentation
	_ "golang.binggl.net/monorepo/mydms/docs"
	// include the mysql driver for runtime
	_ "github.com/go-sql-driver/mysql"
	sec "golang.binggl.net/monorepo/pkg/security"
	srv "golang.binggl.net/monorepo/pkg/server"
)

// Args is uded to configure the API server
type Args struct {
	HostName   string
	Port       int
	ConfigFile string
}

// Run starts the mydms Server
// @title mydms API
// @version 2.0
// @description This is the API of the mydms application
// @license.name MIT License
// @license.url https://raw.githubusercontent.com/bihe/mydms-go/master/LICENSE
func Run(version, build string) error {
	api, addr, appConfig := setupAPIServer(version, build)

	// Start server
	go func() {
		srv.PrintServerBanner("mydms", version, build, string(appConfig.Environment), addr)
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
		return err
	}
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

func setupAPIServer(version, build string) (*echo.Echo, string, config.AppConfig) {
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
	v := mydms.VersionInfo{
		Version: version,
		Build:   build,
	}
	if err := registerRoutes(e, con, c, v, l); err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	// enable swagger for API endpoints
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e, fmt.Sprintf("%s:%d", hostName, port), c
}
