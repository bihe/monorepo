package server

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// PrintServerBanner put some nice emojis on the console
func PrintServerBanner(name, version, build, env, addr string) {
	fmt.Printf("%s Starting server '%s'\n", "🚀", name)
	fmt.Printf("%s Version: '%s-%s'\n", "🔖", version, build)
	fmt.Printf("%s Environment: '%s'\n", "🌍", env)
	fmt.Printf("%s Listening on '%s'\n", "💻", addr)
	fmt.Printf("%s Ready!\n", "🏁")
}

// RunOptions are used to startup a http.Server
type RunOptions struct {
	HostName      string
	Port          int
	AppName       string
	Version       string
	Build         string
	Environment   string
	ServerHandler http.Handler
	Logger        logging.Logger
}

// Run configures and starts the Server
func Run(opt RunOptions) (err error) {
	addr := fmt.Sprintf("%s:%d", opt.HostName, opt.Port)
	httpSrv := &http.Server{Addr: addr, Handler: opt.ServerHandler}
	go func() {
		PrintServerBanner(opt.AppName, opt.Version, opt.Build, opt.Environment, httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			return
		}
	}()
	return Graceful(httpSrv, 5*time.Second, opt.Logger)
}

// Graceful is used to shutdown a server in a graceful manner
func Graceful(s *http.Server, timeout time.Duration, logger logging.Logger) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	logger.Info(fmt.Sprintf("Shutdown with timeout: %s", timeout))
	if err := s.Shutdown(ctx); err != nil {
		return err
	}
	logger.Info("Server stopped")
	return nil
}

// ReadConfig parses supplied application parameters and reads the application config file
func ReadConfig[T any](envPrefix string) (hostname string, port int, basePath string, conf T) {
	flag.String("hostname", "localhost", "the server hostname")
	flag.Int("port", 3000, "network port to listen")
	flag.String("basepath", "./", "the base path of the application")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	v := viper.NewWithOptions(viper.KeyDelimiter("__"))
	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		panic(fmt.Sprintf("Could not bind to command line: %v", err))
	}

	basePath = v.GetString("basepath")
	hostname = v.GetString("hostname")
	port = v.GetInt("port")

	v.SetConfigName("application")                  // name of config file (without extension)
	v.SetConfigType("yaml")                         // type of the config-file
	v.AddConfigPath(path.Join(basePath, "./_etc/")) // path to look for the config file in
	v.AddConfigPath(path.Join(basePath, "./etc/"))  // path to look for the config file in
	v.AddConfigPath(path.Join(basePath, "."))       // optionally look for config in the working directory
	v.SetEnvPrefix(envPrefix)                       // use this prefix for environment variables to overwrite
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Could not get server configuration values: %v", err))
	}
	var c T
	if err := v.Unmarshal(&c); err != nil {
		panic(fmt.Sprintf("Could not unmarshal server configuration values: %v", err))
	}
	conf = c
	return
}

// SetupBasicRouter configures typically used middleware components
func SetupBasicRouter(basePath string, cookieSettings config.ApplicationCookies, corsConfig config.CorsSettings, assets config.AssetSettings, logger logging.Logger) chi.Router {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	//r.Use(handler.NewLoggerMiddleware(logger).LoggerContext)
	r.Use(handler.NewRequestLogger(logger).LoggerContext)
	// use the default list of "compressable" content-type
	r.Use(middleware.NewCompressor(5).Handler)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// setup cors for single frontend
	cors := cors.New(cors.Options{
		AllowedOrigins:   corsConfig.Origins,
		AllowedMethods:   corsConfig.Methods,
		AllowedHeaders:   corsConfig.Headers,
		AllowCredentials: corsConfig.Credentials,
		MaxAge:           corsConfig.MaxAge,
	})
	r.Use(cors.Handler)

	if assets.AssetDir != "" {
		// serving static content
		handler.ServeStaticFile(r, "/favicon.ico", filepath.Join(basePath, assets.AssetDir, "favicon.ico"))
		handler.ServeStaticDir(r, assets.AssetPrefix, http.Dir(filepath.Join(basePath, assets.AssetDir)))
	}
	return r
}

// SetupSecureAPIRouter wires the JWT auth for this router
func SetupSecureAPIRouter(errorPath string, jwtOptions config.Security, cookieSettings config.ApplicationCookies, logger logging.Logger) chi.Router {
	apiRouter := chi.NewRouter()
	apiRouter.Use(security.NewJwtMiddleware(security.JwtOptions{
		CacheDuration: jwtOptions.CacheDuration,
		CookieName:    jwtOptions.CookieName,
		ErrorPath:     errorPath,
		JwtIssuer:     jwtOptions.JwtIssuer,
		JwtSecret:     jwtOptions.JwtSecret,
		RedirectURL:   jwtOptions.LoginRedirect,
		RequiredClaim: security.Claim{
			Name:  jwtOptions.Claim.Name,
			URL:   jwtOptions.Claim.URL,
			Roles: jwtOptions.Claim.Roles,
		},
	}, logger).JwtContext)

	return apiRouter
}
