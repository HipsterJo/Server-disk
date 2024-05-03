package controller_auth

import (
	"disk-server/internal/config"
	login_dto "disk-server/internal/http-server/dto"
	loginHandler "disk-server/internal/http-server/handlers/auth/login"
	registerHandler "disk-server/internal/http-server/handlers/auth/register"
	"disk-server/internal/lib/api/response"
	"disk-server/internal/lib/entities"
	"disk-server/internal/storage"
	"encoding/json"
	"log/slog"
	"net/http"
)

func Register(log *slog.Logger, cfg *config.Config, db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u entities.User
		json.NewDecoder(r.Body).Decode(&u)

		resp := registerHandler.Register(log, cfg, u, db)

		response.SendJSONResponse(w, 200, resp)
	}

}

func Login(log *slog.Logger, cfg *config.Config, db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u login_dto.LoginDto
		json.NewDecoder(r.Body).Decode(&u)

		resp := loginHandler.Login(log, cfg, u, db)

		response.SendJSONResponse(w, 200, resp)
	}

}
