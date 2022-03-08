package apiserver

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/app/model"
	"BIP_backend/internal/service/auth"
)

var (
	incorrectUsernameOrPassword = errors.New("incorrect username or password")
)

func (s *Server) handleUserCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var u = &model.User{}
		if err := c.ShouldBindJSON(u); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, responseUserCreate(false))
			return
		}

		store, _ := s.GetStore()
		if err := store.User().Create(u); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, responseUserCreate(false))
			return
		}
		c.JSON(http.StatusOK, responseUserCreate(true))
	}
}

func (s *Server) handleSessionsCreate() gin.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(c *gin.Context) {
		var r = &request{}
		if err := c.ShouldBindJSON(r); err != nil {
			c.AbortWithError(http.StatusUnauthorized, incorrectUsernameOrPassword)
			return
		}

		store, _ := s.GetStore()
		u, err := store.User().FindByUsername(r.Username)
		if err != nil || !u.ComparePassword(r.Password) {
			c.AbortWithError(http.StatusUnauthorized, incorrectUsernameOrPassword)
			return
		}

		configAuth, err := auth.NewConfig()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		authorizer := auth.NewAuthorizer(configAuth)
		jwt, err := authorizer.GenerateToken(u)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, responseSessionsCreate(jwt, u))
	}
}

func (s *Server) handleTestAuth() gin.HandlerFunc {
	// temporarily for testing
	return func(c *gin.Context) {
		id, ok := c.Get("userID")
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}

func responseUserCreate(success bool) gin.H {
	return gin.H{"success": success}
}

func responseSessionsCreate(jwt string, user *model.User) gin.H {
	return gin.H{
		"jwt":  jwt,
		"user": user,
	}
}
