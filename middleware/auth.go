package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/service/auth"
)

var (
	emptyToken                     = errors.New("empty token")
	incorrectToken                 = errors.New("incorrect token")
	accountTypeDoesNotMatchRequest = errors.New("account type does not match the request")
)

type errorResponse struct {
	Error string `json:"error"`
}

func UserIdentityWithUnauthorizedToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		userID, _, authorized, err := getTokenInfo(token)
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}
		if authorized {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken)
			c.Abort()
			return
		}
		setParams(c, userID, authorized)
	}
}

func UserIdentityWithAuthorizedToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		userID, _, authorized, err := getTokenInfo(token)
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}
		if !authorized {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken)
			c.Abort()
			return
		}
		setParams(c, userID, authorized)
	}
}

func PhotographerIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		_, isPhotographer, _, err := getTokenInfo(token)
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}
		if !isPhotographer {
			newErrorResponse(c, http.StatusBadRequest, accountTypeDoesNotMatchRequest)
			c.Abort()
			return
		}
	}
}

func OrdinaryUserIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		_, isPhotographer, _, err := getTokenInfo(token)
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}
		if isPhotographer {
			newErrorResponse(c, http.StatusBadRequest, accountTypeDoesNotMatchRequest)
			c.Abort()
			return
		}
	}
}

func getTokenInfo(token string) (int /*user id*/, bool /*is photographer*/, bool /*authorizer*/, error) {
	if token == "" {
		return 0, false, false, emptyToken
	}

	authorizer, err := auth.NewAuthorizer()
	if err != nil {
		return 0, false, false, incorrectToken
	}

	userID, isPhotographer, authorized, err := authorizer.ParseToken(token)
	if err != nil {
		return 0, false, false, incorrectToken
	}

	return userID, isPhotographer, authorized, nil
}

func newErrorResponse(c *gin.Context, httpError int, definition error) {
	c.JSON(httpError, errorResponse{Error: definition.Error()})
}

func setParams(c *gin.Context, userID int, authorized bool) {
	c.Set("user_id", userID)
	c.Set("authorized", authorized)
}
