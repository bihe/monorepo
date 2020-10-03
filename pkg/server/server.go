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

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/handler"
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
	return Graceful(httpSrv, 5*time.Second)
}

// Graceful is used to shutdown a server in a graceful manner
func Graceful(s *http.Server, timeout time.Duration) error {
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

// ReadConfig parses supplied application parameters and reads the application config file
func ReadConfig(envPrefix string, getCfg func() interface{}) (hostname string, port int, basePath string, conf interface{}) {
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
	viper.SetEnvPrefix(envPrefix)                       // use this prefix for environment variabls to overwrite
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Could not get server configuration values: %v", err))
	}
	c := getCfg()
	if err := viper.Unmarshal(c); err != nil {
		panic(fmt.Sprintf("Could not unmarshall server configuration values: %v", err))
	}
	conf = c
	return
}

// SetupBasicRouter configures typically used middleware components
func SetupBasicRouter(basePath string, cookieSettings config.ApplicationCookies, corsConfig config.CorsSettings, assets config.AssetSettings, logger *log.Entry) chi.Router {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(handler.NewLoggerMiddleware(logger).LoggerContext)
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

	// serving static content
	handler.ServeStaticFile(r, "/favicon.ico", filepath.Join(basePath, assets.AssetDir, "favicon.ico"))
	handler.ServeStaticDir(r, assets.AssetPrefix, http.Dir(filepath.Join(basePath, assets.AssetDir)))

	return r
}

// SetupSecureAPIRouter wires the JWT auth for this router
func SetupSecureAPIRouter(errorPath string, jwtOptions config.Security, cookieSettings config.ApplicationCookies, logger *log.Entry) chi.Router {
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
	}, cookies.Settings{
		Domain: cookieSettings.Domain,
		Path:   cookieSettings.Path,
		Prefix: cookieSettings.Prefix,
		Secure: cookieSettings.Secure,
	}, logger).JwtContext)

	return apiRouter
}
