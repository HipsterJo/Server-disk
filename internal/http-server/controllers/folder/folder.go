package folder

import (
	"disk-server/internal/config"
	"disk-server/internal/http-server/dto"
	"disk-server/internal/lib/api/response"
	jwt_token "disk-server/internal/lib/jwt"
	"disk-server/internal/storage"
	"encoding/json"
	"log/slog"
	"net/http"
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
		storage.CreateFolder(db, folder, user)

	}
}
