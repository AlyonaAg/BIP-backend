package apiserver

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/app/model"
	"BIP_backend/internal/service/auth"
	"BIP_backend/internal/service/mail"
	"BIP_backend/middleware"
)

var (
	incorrectUsernameOrPassword = errors.New("incorrect username or password")
	incorrectCode               = errors.New("incorrect code")
)

func (s *Server) handleUserCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var u = &model.User{}
		if err := c.ShouldBindJSON(u); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, responseUserCreate(false))
			return
		}

		store, err := s.GetStore()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

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

		store, err := s.GetStore()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		u, err := store.User().FindByUsername(r.Username)
		if err != nil || !u.ComparePassword(r.Password) {
			c.AbortWithError(http.StatusUnauthorized, incorrectUsernameOrPassword)
			return
		}

		authorizer, err := auth.NewAuthorizer()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		jwt, err := authorizer.GenerateToken(u, false /* authorized */)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		code, err := authorizer.GeneratePassword()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		sender, err := mail.NewSender()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// не знаю насколько тут уместно использовать горутину
		// с одной стороны ошибку надо обработать, но с другой нужно ответ отправить быстрее
		// (иначе оно с задержкой небольшой отправляет)
		// отпишись, что думаешь по этому поводу
		go sender.SendMail(u.Mail, code)

		cache, err := s.GetCache()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if err := cache.Set(u.Username, code); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, responseSessionsCreate(jwt))
	}
}

func (s *Server) handler2Factor() gin.HandlerFunc {
	type request struct {
		Code string `json:"code"`
	}

	return func(c *gin.Context) {
		var r = &request{}
		if err := c.ShouldBindJSON(r); err != nil {
			c.AbortWithError(http.StatusUnauthorized, incorrectUsernameOrPassword)
			return
		}

		handleUserIdentity := middleware.UserIdentity()
		handleUserIdentity(c)

		username, ok := c.Get("username")
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		authorized, ok := c.Get("authorized")
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		cache, err := s.GetCache()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		code, err := cache.Get(username.(string))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if code != r.Code {
			c.AbortWithError(http.StatusUnauthorized, incorrectCode)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		u, err := store.User().FindByUsername(username.(string))
		if err != nil || authorized.(bool) {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		authorizer, err := auth.NewAuthorizer()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		jwt, err := authorizer.GenerateToken(u, true /* authorized */)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		cache.Del(username.(string))
		c.JSON(http.StatusOK, response2Factor(jwt, u))
	}
}

func (s *Server) handleTestAuth() gin.HandlerFunc {
	// temporarily for testing
	return func(c *gin.Context) {
		username, ok := c.Get("username")
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authorized, ok := c.Get("authorized")
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.JSON(http.StatusOK, gin.H{"username": username,
			"authorized": authorized})
	}
}

func responseUserCreate(success bool) gin.H {
	return gin.H{"success": success}
}

func responseSessionsCreate(jwt string) gin.H {
	return gin.H{
		"jwt": jwt,
	}
}

func response2Factor(jwt string, user *model.User) gin.H {
	return gin.H{
		"jwt":  jwt,
		"user": user,
	}
}
