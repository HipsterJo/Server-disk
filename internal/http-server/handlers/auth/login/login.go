package login

import (
	"disk-server/internal/config"
	login_dto "disk-server/internal/http-server/dto"
	"disk-server/internal/lib/api/response"
	hashassword "disk-server/internal/lib/hashPassword"
	jwt_token "disk-server/internal/lib/jwt"
	"disk-server/internal/storage"
	"log/slog"
)

func Login(log *slog.Logger, cfg *config.Config, user login_dto.LoginDto, db *storage.Storage) response.Response {
	u, resp := storage.GetUserByUserName(db, user.Username)
	if resp.Error != "" {
		log.Error("Wrong data")
		return response.Error("Wrong data")
	}

	isCorrectPassword := hashassword.CheckPasswordHash(user.Password, u.Password)

	if !isCorrectPassword {

		return response.Error("Wrong data")
	}

	validToken, err := jwt_token.GenreateJWT(u)

	if err != nil {
		return response.Error("Wrong data")
	}

	return response.OKWithData(map[string]interface{}{
		"validToken": validToken,
	})
}
