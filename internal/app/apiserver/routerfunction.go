package apiserver

import (
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

		store := s.GetStore()
		if err := store.User().Create(u); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, responseUserCreate(false))
			return
		}
		c.JSON(http.StatusOK, responseUserCreate(true))
	}
}

func responseUserCreate(success bool) gin.H {
	return gin.H{"success": success}
}
