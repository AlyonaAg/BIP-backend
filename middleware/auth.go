package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"

	"BIP_backend/internal/service/auth"
)

var (
	emptyToken     = errors.New("empty token")
	incorrectToken = errors.New("incorrect token")
)

type errorResponse struct {
	Error string `json:"error"`
}

func UserIdentityWithUnauthorizedToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		username, authorized, err := getTokenInfo(token)
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, err)
			c.Abort()
		}
		if authorized {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken)
			c.Abort()
		}
		setParams(c, username, authorized)
	}
}

func UserIdentityWithAuthorizedToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		username, authorized, err := getTokenInfo(token)
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, err)
			c.Abort()
		}
		if !authorized {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken)
			c.Abort()
		}
		setParams(c, username, authorized)
	}
}

func getTokenInfo(token string) (string, bool, error) {
	if token == "" {
		return "", false, emptyToken
	}

	authorizer, err := auth.NewAuthorizer()
	if err != nil {
		return "", false, incorrectToken
	}

	username, authorized, err := authorizer.ParseToken(token)
	if err != nil {
		return "", false, incorrectToken
	}

	return username, authorized, nil
}

func newErrorResponse(c *gin.Context, httpError int, definition error) {
	c.JSON(httpError, errorResponse{Error: definition.Error()})
}

func setParams(c *gin.Context, username string, authorized bool) {
	c.Set("username", username)
	c.Set("authorized", authorized)
}
