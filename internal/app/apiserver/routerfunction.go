package apiserver

import (
	"BIP_backend/internal/app/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (s *Server) handleUserCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := &model.User{}
		if err := c.ShouldBindJSON(u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false})
			return
		}
		if err := s.store.User().Create(u); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
