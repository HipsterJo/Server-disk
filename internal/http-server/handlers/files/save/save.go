package save

import (
	"disk-server/internal/config"
	"disk-server/internal/lib/api/response"
	createfiles "disk-server/internal/lib/createFiles"
	"disk-server/internal/lib/entities"
	jwt_token "disk-server/internal/lib/jwt"
	"disk-server/internal/storage"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
)

type FileSaver interface {
	SaveFile()
}

// Разобораться с return, поч ничего не возращают
func New(log *slog.Logger, cfg *config.Config, s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.files.New"

		log := log.With(slog.String("op", op))

		file, handler, err := r.FormFile("file")
		path := r.FormValue("path")
		size := handler.Size
		mime_type := handler.Header.Values("Content-type")
		contentDisposition := handler.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err != nil {

		}

		filename := params["filename"]
		if filename == "" {
			response.SendJSONResponse(w, 403, response.Error("Нет файла"))
		}

		if err != nil {
			log.Error("Ошибка получения файла", err)
			return
		}

		defer file.Close()

		uniqueName := createfiles.GenerateUniqueFilename()
		//
		var folder_id int
		var access_level int

		if r.Body == nil {
			folder_id = 0
			access_level = 0
		}

		err = json.NewDecoder(r.Body).Decode(&folder_id)
		if err != nil {
			folder_id = 0
		}

		//
		newFileData := entities.FileData{
			FileName:    filename,
			Size:        int(size),
			MimeType:    mime_type[0],
			Path:        uniqueName,
			FolderId:    folder_id,
			AccessLevel: access_level,
		}

		userData, res := jwt_token.GetJsonJwt(r)
		if res.Error != "" {
			response.SendJSONResponse(w, 405, res)
		}

		storage.NewFile(s, userData.Username, newFileData)

		f, err := os.Create(cfg.Upload_path + path + uniqueName)

		if err != nil {
			http.Error(w, "Ошибка при создании файла", http.StatusInternalServerError)
			return
		}

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
