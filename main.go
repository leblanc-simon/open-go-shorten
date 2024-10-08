package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ilyakaznacheev/cleanenv"
	"open-go-shorten.eu/config"
	"open-go-shorten.eu/handlers"
	"open-go-shorten.eu/middleware"
	"open-go-shorten.eu/utils"
)

//go:embed all:static
var staticFs embed.FS

var version = "develop"
var appName = "OpenGoShorten"

type Args struct {
	ConfigPath string
}

var cfg config.Config

func main() {
	args := ProcessArgs(&cfg)
	// read configuration from the file and environment variables
	if _, err := os.Stat(args.ConfigPath); errors.Is(err, os.ErrNotExist) {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	} else {
		if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	utils.InitRedis(&cfg)
	middleware.InitJwt(&cfg)
	handlers.InitAuth(&cfg)

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.SetResponseHeaders(w)
		http.ServeFileFS(w, r, staticFs, "static/index.html")
	}).Methods("GET")

	router.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFs, strings.TrimLeft(r.RequestURI, "/"))
	}).Methods("GET")

	router.HandleFunc("/login", handlers.Login).Methods("POST")

	router.Handle("/shorten", middleware.JWTMiddleware(http.HandlerFunc(handlers.ShortenURL))).Methods("POST")
	router.Handle("/urls", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetURLs))).Methods("GET")
	router.Handle("/{shortURL}", middleware.JWTMiddleware(http.HandlerFunc(handlers.DeleteUrl))).Methods("DELETE")

	router.HandleFunc("/{shortURL}", handlers.RedirectURL).Methods("GET")

	log.Printf("Running server %s", cfg.Server.Host+":"+strconv.Itoa(cfg.Server.Port))
	log.Printf("%s running on version %s", appName, version)

	http.ListenAndServe(cfg.Server.Host+":"+strconv.Itoa(cfg.Server.Port), router)
}

func ProcessArgs(cfg interface{}) Args {
	var a Args

	f := flag.NewFlagSet("OpenGoShorten", 1)

	f.StringVar(&a.ConfigPath, "c", "config.yaml", "Path to configuration file")

	fu := f.Usage
	f.Usage = func() {
		fu()
		envHelp, _ := cleanenv.GetDescription(cfg, nil)
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), envHelp)
	}

	f.Parse(os.Args[1:])

	return a
}
