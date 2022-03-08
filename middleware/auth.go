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

		configAuth, err := auth.NewConfig()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		authorizer := auth.NewAuthorizer(configAuth)
		id, err := authorizer.ParseToken(token)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		c.Set("userID", id)
	}
}
