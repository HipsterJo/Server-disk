package controller_folder

import (
	"disk-server/internal/config"
	"disk-server/internal/http-server/dto"
	"disk-server/internal/lib/api/response"
	"disk-server/internal/lib/conversion"
	jwt_token "disk-server/internal/lib/jwt"
	"disk-server/internal/storage"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func CreateFolder(log *slog.Logger, cfg *config.Config, db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var folder dto.CreateFolder
		json.NewDecoder(r.Body).Decode(&folder)

		userObject, resp := jwt_token.GetJsonJwt(r)

		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		user, resp := storage.GetUserByUserName(db, userObject.Username)
		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		resp = storage.CreateFolder(db, folder, user)

		response.SendJSONResponse(w, 200, resp)

	}
}

func GetFolder(log *slog.Logger, cfg *config.Config, db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")

		id := conversion.StringToInt(idStr, 0)

		userObject, resp := jwt_token.GetJsonJwt(r)

		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		user, resp := storage.GetUserByUserName(db, userObject.Username)
		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		resp = storage.GetFilesFromFolder(db, user, r, id)

		response.SendJSONResponse(w, 200, resp)
	}
}
