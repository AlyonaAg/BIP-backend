package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sethvargo/go-password/password"

	"BIP_backend/internal/app/model"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserID         int
	IsPhotographer bool
	Authorized     bool
}

type Authorizer struct {
	config *Config
}

func NewAuthorizer() (*Authorizer, error) {
	configAuth, err := NewConfig()
	if err != nil {
		return nil, err
	}

	return &Authorizer{
		config: configAuth,
	}, nil
}

func (a *Authorizer) GenerateToken(u *model.User, authorized bool) (string, error) {
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
			u.IsPhotographer,
			authorized,
		})

	signedToken, err := token.SignedString([]byte(config.signingKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (a *Authorizer) ParseToken(tokenString string) (int /*user id*/, bool, /*is photographer*/
	bool /*authorized*/, error) {
	config, err := a.GetConfig()
	if err != nil {
		return 0, false, false, err
	}

	token, err := jwt.ParseWithClaims(tokenString,
		&tokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(config.signingKey), nil
		})
	if err != nil {
		return 0, false, false, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return 0, false, false, errors.New("invalid token")
	}

	return claims.UserID, claims.IsPhotographer, claims.Authorized, nil
}

func (a *Authorizer) GeneratePassword() (string, error) {
	password, err := password.Generate(6, 6, 0, false, false)
	if err != nil {
		return "", err
	}
	return password, err
}

func (a *Authorizer) GetConfig() (*Config, error) {
	if a.config == nil {
		return nil, errors.New("empty auth config")
	}
	return a.config, nil
}
