package apiserver

import (
	"BIP_backend/internal/app/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) handleRegistration() gin.HandlerFunc {
	return func(c *gin.Context) {
		var rt model.User
		if err := c.ShouldBindJSON(&rt); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false})
			return
		}
		fmt.Println(rt)
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
