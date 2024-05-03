package checkAuthMiddleware

import (
	"net/http"
	"strings"

	"disk-server/internal/lib/api/response"
	jwt_token "disk-server/internal/lib/jwt"
)

// CheckAuthMiddleware проверяет наличие токена авторизации в заголовке запроса
func CheckAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			response.SendJSONResponse(w, http.StatusUnauthorized, response.Error("Unauthorized"))
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			response.SendJSONResponse(w, http.StatusUnauthorized, response.Error("Unauthorized"))
			return
		}

		token := tokenParts[1]
		if !jwt_token.CheckJwt(token) {
			response.SendJSONResponse(w, http.StatusUnauthorized, response.Error("Unauthorized"))
			return
		}

		next.ServeHTTP(w, r)
	}
}
