package jwt_token

import (
	"disk-server/internal/lib/api/response"
	"disk-server/internal/lib/entities"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

var secret_key = []byte("key")

func GenreateJWT(user entities.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(time.Hour * 1000).Unix()
	claims["username"] = user.Username
	claims["email"] = user.Email

	tokenString, err := token.SignedString(secret_key)

	if err != nil {
		log.Fatal(err)
	}

	return tokenString, nil

}

func CheckJwt(jwtToken string) bool {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret_key, nil
	})

	if err != nil {
		log.Println("Failed to parse token:", err)
		return false
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true
	}

	return false
}

func GetJsonJwt(r *http.Request) (UserData, response.Response) {
	authHeader := r.Header.Get("Authorization")

	authParts := strings.Split(authHeader, " ")
	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return UserData{}, response.Error("Ошибка авторизации")
	}
	tokenString := authParts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret_key, nil
	})
	if err != nil || !token.Valid {
		return UserData{}, response.Error("Ошибка авторизации")
	}

	claims := token.Claims.(jwt.MapClaims)

	var userData UserData

	userData.Username = claims["username"].(string)
	userData.Email = claims["email"].(string)

	return userData, response.OK()
}
