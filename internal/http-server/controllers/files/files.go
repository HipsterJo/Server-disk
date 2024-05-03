package files

import (
	"disk-server/internal/config"
	"disk-server/internal/lib/api/response"
	jwt_token "disk-server/internal/lib/jwt"
	"disk-server/internal/storage"
	"log/slog"
	"net/http"
	"strings"
)

func GetUserFiles(log *slog.Logger, cfg *config.Config, db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			response.SendJSONResponse(w, 405, response.Error("Ошибка авторизации"))
		}
		tokenString := authParts[1]

		userObject, err := jwt_token.GetJsonJwt(tokenString)

		if err != nil {
			response.SendJSONResponse(w, 405, response.Error("Ошибка авторизации"))
		}

		storage.GetUserByUserName(db, userObject.Username)
	}
}
