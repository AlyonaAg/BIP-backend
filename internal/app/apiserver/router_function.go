package apiserver

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func (s *Server) handleRegistration() gin.HandlerFunc {
	type Request struct {
		First_name      string `json:"first_name"`
		Second_name     string `json:"second_name"`
		Is_photographer bool   `json:"is_photographer"`
		Avatar_URL      string `json:"avatar_url"`
		Phone_number    string `json:"phone_number"`
		Mail            string `json:"mail"`
	}

	return func(c *gin.Context) {
		var rt Request
		if err := c.BindJSON(&rt); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func (s *Server) handleHello() gin.HandlerFunc {
	//for local var
	return func(c *gin.Context) {
		io.WriteString(c.Writer, "Hello")
	}
}
