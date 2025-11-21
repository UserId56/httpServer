package services

import (
	"fmt"
	"httpServer/models"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(user *models.User) (string, error) {
	configTime := os.Getenv("SERVER_JWT_EXPIRE_TIME")

	minutes, err := strconv.Atoi(configTime)
	if err != nil || minutes <= 0 {
		minutes = 600
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
	},
		jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ValidationFields(userData map[string]interface{}) error {
	type validationConfig struct {
		typeValue string
		rules     []string
	}
	calidationRules := map[string]validationConfig{
		"username": {
			typeValue: "text",
			rules:     []string{"required", "min=3", "max=50"},
		},
		"email": {
			typeValue: "text",
			rules:     []string{"required", "email"},
		},
		"password": {
			typeValue: "text",
			rules:     []string{"required", "min=8"},
		},
		"avatar": {
			typeValue: "text",
			rules:     []string{},
		},
		"bio": {
			typeValue: "text",
			rules:     []string{},
		},
		"role_id": {
			typeValue: "uint",
			rules:     []string{},
		},
	}
	var fieldsList []string
	for field := range userData {
		fieldsList = append(fieldsList, field)
	}
	for _, field := range fieldsList {
		config, exists := calidationRules[field]
		if !exists {
			return fmt.Errorf("неизвестное поле: %s", field)
		}
		value := userData[field]
		validate := validator.New()
		switch config.typeValue {
		case "text":
			strValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("поле %s должно быть строкой", field)
			}
			for _, rule := range config.rules {
				err := validate.Var(strValue, rule)
				if err != nil {
					return fmt.Errorf("поле %s не валидно", field)
				}
			}
		}
	}
	return nil
}
