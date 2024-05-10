package files

import (
	"disk-server/internal/config"
	"disk-server/internal/http-server/dto"
	response "disk-server/internal/lib/api/response"
	jwt_token "disk-server/internal/lib/jwt"
	"disk-server/internal/storage"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func GetUserFiles(log *slog.Logger, cfg *config.Config, db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userObject, resp := jwt_token.GetJsonJwt(r)

		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		user, resp := storage.GetUserByUserName(db, userObject.Username)
		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		params := storage.InitialFilter(r)
		files, resp := storage.GetUserFiles(db, user, params, r)

		response.SendJSONResponse(w, 200, response.OKWithData(files))
	}
}
func GetHomeDir(log *slog.Logger, cfg *config.Config, db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userObject, resp := jwt_token.GetJsonJwt(r)

		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		user, resp := storage.GetUserByUserName(db, userObject.Username)
		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		resp = storage.GetHomeDir(db, user, r)

		response.SendJSONResponse(w, 200, resp)
	}
}

func AddFileToFolder(log *slog.Logger, cfg *config.Config, db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dto dto.FileToFolder
		err := json.NewDecoder(r.Body).Decode(&dto)
		fmt.Println(err)
		userObject, resp := jwt_token.GetJsonJwt(r)

		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		user, resp := storage.GetUserByUserName(db, userObject.Username)
		if resp.Error != "" {
			response.SendJSONResponse(w, 405, resp)
		}

		resp = storage.AddFileToFolder(db, dto.FolderId, dto.FildeId, user)

		response.SendJSONResponse(w, 200, resp)
	}
}
