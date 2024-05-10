package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"disk-server/internal/config"

	controller_auth "disk-server/internal/http-server/controllers/auth"
	controller_files "disk-server/internal/http-server/controllers/files"
	controller_folder "disk-server/internal/http-server/controllers/folder"
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

	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.Upload_path))))
	router.Post("/register", controller_auth.Register(log, cfg, storage))
	router.Post("/login", controller_auth.Login(log, cfg, storage))
	router.Post("/upload", checkAuthMiddleware.CheckAuthMiddleware(save.New(log, cfg, storage)))
	router.Get("/files", checkAuthMiddleware.CheckAuthMiddleware(controller_files.GetUserFiles(log, cfg, storage)))
	router.Get("/home", checkAuthMiddleware.CheckAuthMiddleware(controller_files.GetHomeDir(log, cfg, storage)))
	router.Post("/folder", checkAuthMiddleware.CheckAuthMiddleware(controller_folder.CreateFolder(log, cfg, storage)))
	router.Get("/folder/{id}", checkAuthMiddleware.CheckAuthMiddleware(controller_folder.GetFolder(log, cfg, storage)))
	router.Post("/addFileToFolder", checkAuthMiddleware.CheckAuthMiddleware(controller_files.AddFileToFolder(log, cfg, storage)))
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
