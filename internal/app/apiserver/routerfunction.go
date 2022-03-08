package apiserver

import (
	"BIP_backend/internal/service/auth"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/app/model"
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
			c.JSON(http.StatusUnauthorized, errorSessionsCreate().Error())
			return
		}

		store, _ := s.GetStore()
		u, err := store.User().FindByUsername(r.Username)
		if err != nil || !u.ComparePassword(r.Password) {
			c.JSON(http.StatusUnauthorized, errorSessionsCreate().Error())
			return
		}

		configAuth, err := auth.NewConfig()
		if err != nil {
			c.JSON(http.StatusInternalServerError, "")
			return
		}

		authorizer := auth.NewAuthorizer(configAuth)
		jwt, err := authorizer.GenerateToken(u)
		if err != nil {
			c.JSON(http.StatusUnauthorized, errorSessionsCreate().Error())
			return
		}

		c.JSON(http.StatusOK, responseSessionsCreate(jwt, u))
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

func errorSessionsCreate() error {
	return errors.New("incorrect username or password")
}
