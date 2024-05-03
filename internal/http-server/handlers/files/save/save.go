package save

import (
	"disk-server/internal/config"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

type FileSaver interface {
	SaveFile()
}

func ensureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func New(log *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.files.New"

		log := log.With(slog.String("op", op))

		file, handler, err := r.FormFile("file")
		path := r.FormValue("path")
		if err != nil {
			log.Error("Ошибка получения файла", err)
			return
		}

		defer file.Close()

		// Ensure directory exists
		if err := ensureDirExists(cfg.Upload_path + path); err != nil {
			log.Error("Ошибка создания папки", err)
			http.Error(w, "Ошибка при создании папки", http.StatusInternalServerError)
			return
		}

		f, err := os.Create(cfg.Upload_path + path + handler.Filename)

		if err != nil {
			http.Error(w, "Ошибка при создании файла", http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, "Ошибка при сохранении файла", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Файл %s успешно загружен\n", handler.Filename)

	}
}
