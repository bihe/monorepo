package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/bihe/mydms/internal"
	"github.com/bihe/mydms/internal/config"
	"github.com/bihe/mydms/internal/errors"
	"github.com/bihe/mydms/internal/persistence"
	"github.com/bihe/mydms/internal/security"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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

func parseFlags() *ServerArgs {
	c := new(ServerArgs)
	flag.StringVar(&c.HostName, "hostname", "localhost", "the server hostname")
	flag.IntVar(&c.Port, "port", 3000, "network port to listen")
	flag.StringVar(&c.ConfigFile, "c", "application.json", "path to the application c file")
	flag.Parse()
	return c
}

func configFromFile(configFileName string) config.AppConfig {
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

func setupAPIServer() (*echo.Echo, string) {
	args := parseFlags()
	c := configFromFile(args.ConfigFile)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = errors.CustomErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	InitLogger(c.Log, e)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     c.Cors.AllowedOrigins,
		AllowHeaders:     c.Cors.AllowedHeaders,
		AllowMethods:     c.Cors.AllowedMethods,
		AllowCredentials: c.Cors.AllowCredentials,
		MaxAge:           c.Cors.MaxAge,
	}))

	e.Use(middleware.Secure())
	e.Use(security.JwtWithConfig(security.JwtOptions{
		JwtSecret:  c.Sec.JwtSecret,
		JwtIssuer:  c.Sec.JwtIssuer,
		CookieName: c.Sec.CookieName,
		RequiredClaim: sec.Claim{
			Name:  c.Sec.Claim.Name,
			URL:   c.Sec.Claim.URL,
			Roles: c.Sec.Claim.Roles,
		},
		RedirectURL:   c.Sec.LoginRedirect,
		CacheDuration: c.Sec.CacheDuration,
	}))

	// persistence store && application version
	con := persistence.NewConn(c.DB.ConnStr)
	version := internal.VersionInfo{
		Version: Version,
		Build:   Build,
	}
	if err := registerRoutes(e, con, c, version); err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	// enable swagger for API endpoints
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e, fmt.Sprintf("%s:%d", args.HostName, args.Port)
}
