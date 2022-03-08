package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"

	"BIP_backend/internal/app/model"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int
}

type Authorizer struct {
	config *Config
}

func NewAuthorizer(config *Config) *Authorizer {
	return &Authorizer{
		config: config,
	}
}

func (a *Authorizer) GenerateToken(u *model.User) (string, error) {
	config, err := a.GetConfig()
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		&tokenClaims{
			jwt.StandardClaims{
				IssuedAt: time.Now().Unix(),
				ExpiresAt: time.Now().Add(
					time.Hour * time.Duration(config.expireDuration),
				).Unix(),
			},
			u.ID,
		})

	signedToken, err := token.SignedString([]byte(config.signingKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (a *Authorizer) ParseToken(tokenString string) (int, error) {
	config, err := a.GetConfig()
	if err != nil {
		return 0, err
	}

	token, err := jwt.ParseWithClaims(tokenString,
		&tokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(config.signingKey), nil
		})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	return claims.UserId, nil
}

func (a *Authorizer) GetConfig() (*Config, error) {
	if a.config == nil {
		return nil, errors.New("empty auth config")
	}
	return a.config, nil
}
