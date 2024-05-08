package files

import (
	"disk-server/internal/config"
	response "disk-server/internal/lib/api/response"
	jwt_token "disk-server/internal/lib/jwt"
	"disk-server/internal/storage"
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
		storage.GetUserFiles(db, user, params, r)
	}
}
