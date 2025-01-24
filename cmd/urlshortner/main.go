package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/FacelessWayfarer/urlshortner/internal/config"
	"github.com/FacelessWayfarer/urlshortner/internal/database/sqllite"
	urldelete "github.com/FacelessWayfarer/urlshortner/internal/handlers/url-delete"
	urlget "github.com/FacelessWayfarer/urlshortner/internal/handlers/url-get"
	urlsave "github.com/FacelessWayfarer/urlshortner/internal/handlers/url-save"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/slogg"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting urlshortner", slog.String("env:", cfg.Env))

	db, err := sqllite.New(cfg.DatabasePath)
	if err != nil {
		log.Error("failed to init database", slogg.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	// router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortner", map[string]string{
			cfg.HTTPServ.User: cfg.HTTPServ.Password,
		}))

		r.Post("/", urlsave.New(log, db, db))
		r.Delete("/{alias}", urldelete.New(log, db))
	})

	router.Get("/{alias}", urlget.New(log, db))

	log.Info("starting server", slog.String("address:", cfg.Address))

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServ.Timeout,
		WriteTimeout: cfg.HTTPServ.Timeout,
		IdleTimeout:  cfg.HTTPServ.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server", slogg.Err(err))
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return log
}
