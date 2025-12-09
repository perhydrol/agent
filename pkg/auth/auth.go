package auth

import (
	"errors"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/perhydrol/insurance-agent-backend/pkg/config"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
)

type myCustomClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var secret string

func initJwt() {
	sync.OnceFunc(func() {
		secret = config.AppConfig.Server.JwtPasswd
		if secret == "" {
			if config.AppConfig.Server.Mode == "debug" {
				logger.Log.Warn("failed to read JWT secret, using default")
				secret = "insurai_default_secret_key"
			} else {
				logger.Log.Panic("failed to read JWT secret")
			}
		}
	})
}

func GenerateToken(userID int64, expire time.Duration, username string) (string, error) {
	initJwt()
	// Token 有效期 24 小时
	claims := myCustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			Issuer:    "insurai",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString string) (*myCustomClaims, error) {
	initJwt()

	token, err := jwt.ParseWithClaims(tokenString, &myCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*myCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
