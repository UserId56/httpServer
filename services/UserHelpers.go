package services

import (
	"httpServer/models"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(user *models.User) (string, error) {
	configTime := os.Getenv("SERVER_JWT_EXPIRE_TIME")

	minutes, err := strconv.Atoi(configTime)
	if err != nil || minutes <= 0 {
		minutes = 15
	}
	exp := time.Now().Add(time.Duration(minutes) * time.Minute)
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"role_id":  user.RoleID,
		"username": user.Username,
		"exp":      jwt.NewNumericDate(exp),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := os.Getenv("SERVER_JWT_SECRET_KEY")
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ParseJWT(tokenString string) (*jwt.Token, error) {
	secretKey := os.Getenv("SERVER_JWT_SECRET_KEY")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenMalformed
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
