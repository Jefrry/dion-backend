// @title           Dion API
// @version         1.0
// @host            localhost:8080
// @BasePath        /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	_ "dion-backend/docs"
	"dion-backend/internal/config"
	"dion-backend/internal/db"
	"dion-backend/internal/handler"
	"dion-backend/internal/lib/logger/handlers/slogpretty"
	"dion-backend/internal/repo"
	"dion-backend/internal/router"
	"dion-backend/internal/service"
	"dion-backend/internal/utils"

	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	gormDB := db.MustConnect(cfg.DBConfig, log)

	recordingsRepo := repo.NewRecordingsRepo(gormDB)
	recordingsService := service.NewRecordingsDataService(recordingsRepo)

	artistsRepo := repo.NewArtistsRepo(gormDB)
	artistsService := service.NewArtistsDataService(artistsRepo)

	handlerUtils := utils.NewHandlerUtils()
	recordingsHandler := handler.NewRecordingsHandler(log, handlerUtils, recordingsService)
	artistsHandler := handler.NewArtistsHandler(log, handlerUtils, artistsService)
	adminHandler := handler.NewAdminHandler(log, handlerUtils, cfg.AdminConfig)
	r := router.NewRouter(recordingsHandler, artistsHandler, adminHandler, cfg.AdminConfig).MustRun()

	log.Info("starting server")
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started", "addr", cfg.Address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	loggerHandler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(loggerHandler)
}
