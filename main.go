package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"leblanc.io/open-go-base/appconf"
	"leblanc.io/open-go-base/logx"
	"leblanc.io/open-go-base/ratelimit"
	"open-go-shorten.eu/config"
	"open-go-shorten.eu/handlers"
	"open-go-shorten.eu/middleware"
	"open-go-shorten.eu/utils"
)

//go:embed all:static
var staticFs embed.FS

var version = "develop"
var appName = "OpenGoShorten"

func main() {
	cfg := &config.Config{}
	appconf.MustLoad(cfg, appconf.Options{Name: appName, Version: version})

	logger := logx.New(cfg.Log)
	slog.SetDefault(logger)

	if err := run(cfg, logger); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config, logger *slog.Logger) error {
	limiter, err := ratelimit.New(cfg.Server.RateLimit, cfg.Server.TrustedProxies)
	if err != nil {
		return err
	}

	if err := checkAuth(cfg, logger); err != nil {
		return err
	}

	utils.InitRedis(cfg)
	pingCtx, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	if err := utils.PingRedis(pingCtx); err != nil {
		logger.Warn("Redis is not reachable", "err", err)
	}
	cancelPing()

	middleware.InitJwt(cfg)
	handlers.InitAuth(cfg)
	handlers.InitQRCode(cfg)
	handlers.InitClientIP(limiter.ClientIP)

	staticDir, err := fs.Sub(staticFs, "static")
	if err != nil {
		return err
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		utils.SetResponseHeaders(w)
		http.ServeFileFS(w, r, staticDir, "index.html")
	})
	router.Handle("GET /static/", http.StripPrefix("/static", noDirListing(http.FileServerFS(staticDir))))

	router.HandleFunc("POST /login", handlers.Login)

	router.Handle("POST /shorten", middleware.JWTMiddleware(http.HandlerFunc(handlers.ShortenURL)))
	router.Handle("GET /urls", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetURLs)))
	router.Handle("DELETE /{shortURL}", middleware.JWTMiddleware(http.HandlerFunc(handlers.DeleteUrl)))

	router.HandleFunc("GET /qrcode/{shortURL}/{format}", handlers.QRCode)
	router.HandleFunc("GET /{shortURL}", handlers.RedirectURL)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           limiter.Middleware(router),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "app", appName, "version", version, "addr", addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// oldDefaultJwtSecret was the default JWT secret of the previous versions: being public, it
// allows anyone to forge authentication tokens.
const oldDefaultJwtSecret = "019263ef-1532-7333-98d6-3f08007f99f2"

// checkAuth warns about an insecure authentication configuration, and
// generates a random JWT secret when none is configured.
func checkAuth(cfg *config.Config, logger *slog.Logger) error {
	if cfg.Auth.Password == config.DefaultPasswordHash {
		logger.Warn("the default administrator password is used, change it")
	}

	if cfg.Auth.JwtSecret == oldDefaultJwtSecret {
		cfg.Auth.JwtSecret = ""
		logger.Warn("the JWT secret is the public default of previous versions, it is ignored")
	}

	if cfg.Auth.JwtSecret == "" {
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return err
		}
		cfg.Auth.JwtSecret = hex.EncodeToString(secret)
		logger.Warn("no JWT secret configured, a random one is used: sessions are lost on restart")
	}
	return nil
}

// noDirListing prevents the file server from listing the content of the
// embedded directories.
func noDirListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
