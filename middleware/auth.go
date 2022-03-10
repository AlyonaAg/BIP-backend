package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/service/auth"
)

var (
	emptyToken = errors.New("empty token")
)

func UserIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithError(http.StatusUnauthorized, emptyToken)
			return
		}

		authorizer, err := auth.NewAuthorizer()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		username, authorized, err := authorizer.ParseToken(token)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		c.Set("username", username)
		c.Set("authorized", authorized)
	}
}
