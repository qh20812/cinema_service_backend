package auth

import (
	"main/src/common"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = common.GetEnv("JWT_SECRET")

func GenerateToken(userID string, role string) (string, error){
	claims := jwt.MapClaims{
		"user_id": userID,
		"role": role,
		"exp": time.Now().Add(time.Hour*24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}