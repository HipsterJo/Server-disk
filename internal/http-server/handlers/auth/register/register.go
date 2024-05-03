package register

import (
	"disk-server/internal/config"
	"disk-server/internal/lib/api/response"
	"disk-server/internal/lib/entities"
	hashpassword "disk-server/internal/lib/hashPassword"
	jwt_token "disk-server/internal/lib/jwt"
	"disk-server/internal/storage"
	"log/slog"
)

func Register(log *slog.Logger, cfg *config.Config, user entities.User, db *storage.Storage) response.Response {
	allUsers, err := storage.GetAllUsers(db)
	if err != nil {
		log.Error("Failed to retrieve users from database:", err)
		return response.Error("Что-то пошло не так")
	}

	for _, u := range allUsers {
		if u.Username == user.Username || u.Email == user.Email {
			log.Error("User already exists")
			return response.Error("User already exists")
		}
	}

	hashPass, err := hashpassword.HashPassword(user.Password)

	if err != nil {
		log.Error("Failed to hashPass:", err)
		return response.Error(err.Error())
	}

	user.Password = hashPass

	err = storage.CreateUser(user, db)
	if err != nil {
		log.Error("Failed to create user:", err)
		return response.Error(err.Error())
	}

	validToken, err := jwt_token.GenreateJWT(user)

	if err != nil {
		return response.Error("Wrong data")
	}

	return response.OKWithData(map[string]interface{}{
		"validToken": validToken,
	})
}
