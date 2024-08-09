package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"github.com/zanzhit/flat-seller/internal/config"
	authhandler "github.com/zanzhit/flat-seller/internal/http-server/handlers/auth"
	flathandler "github.com/zanzhit/flat-seller/internal/http-server/handlers/flat"
	househandler "github.com/zanzhit/flat-seller/internal/http-server/handlers/house"
	authmid "github.com/zanzhit/flat-seller/internal/http-server/middleware/auth"
	"github.com/zanzhit/flat-seller/internal/http-server/middleware/logger"
	"github.com/zanzhit/flat-seller/internal/lib/logger/sl"
	authservice "github.com/zanzhit/flat-seller/internal/services/auth"
	flatservice "github.com/zanzhit/flat-seller/internal/services/flat"
	"github.com/zanzhit/flat-seller/internal/storage/postgres"
	authstorage "github.com/zanzhit/flat-seller/internal/storage/postgres/auth"
	flatstorage "github.com/zanzhit/flat-seller/internal/storage/postgres/flat"
	housestorage "github.com/zanzhit/flat-seller/internal/storage/postgres/house"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application", slog.Any("config", cfg))

	cfg.DB.Password = os.Getenv("POSTGRES_PASSWORD")
	if cfg.DB.Password == "" {
		panic("POSTGRES_PASSWORD is required")
	}

	storage, err := postgres.New(*cfg)
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	authstorage := authstorage.New(storage)
	authService := authservice.New(log, authstorage, authstorage, cfg.TokenTTL, cfg.Secret)
	authhandler := authhandler.New(log, authService)

	houseStorage := housestorage.New(storage)
	houseHandler := househandler.New(log, houseStorage)

	flatStorage := flatstorage.New(storage)
	flatService := flatservice.New(log, flatStorage)
	flatHandler := flathandler.New(log, flatService)

	router.Post("/register", authhandler.RegisterNewUser)
	router.Post("/login", authhandler.Login)

	router.With(authmid.JWTAuth(cfg.Secret)).Group(func(r chi.Router) {
		r.Post("/flat/create", flatHandler.SaveFlat)
		r.With(authmid.AdminRequired).Post("/flat/update", flatHandler.UpdateFlat)
		r.With(authmid.AdminRequired).Post("/house/create", houseHandler.SaveHouse)
		r.Get("/house/{id}", houseHandler.House)
	})

	log.Info("starting http server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	<-done
	log.Error("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	if err := storage.Close(); err != nil {
		log.Error("failed to close storage", sl.Err(err))

		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
