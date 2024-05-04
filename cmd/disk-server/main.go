package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"disk-server/internal/config"

	controller_auth "disk-server/internal/http-server/controllers/auth"
	"disk-server/internal/http-server/handlers/files/save"
	checkAuthMiddleware "disk-server/internal/http-server/middleware/checkAuth"
	"disk-server/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	configPath := "./config/config.yaml"
	connStr := "user=postgres password=12345678 dbname=disk-cloud sslmode=disable"
	cfg := config.MustLoad(configPath)
	log := setupLogger(cfg.Env)
	storage, err := storage.New(connStr)

	if err != nil {
		log.Error("Wrong Path", err)
	}

	log.Info("Starting server")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.URLFormat)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "HELLO11")
	})
	router.Post("/register", controller_auth.Register(log, cfg, storage))
	router.Post("/login", controller_auth.Login(log, cfg, storage))
	router.Post("/upload", checkAuthMiddleware.CheckAuthMiddleware(save.New(log, cfg, storage)))
	// router.Get("/files", checkAuthMiddleware.CheckAuthMiddleware(controller_files.GetUserFiles(log, cfg, storage)))

	http.ListenAndServe(":8080", router)
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(slog.NewJSONHandler(
			os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}

	return log
}
