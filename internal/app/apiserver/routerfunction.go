package apiserver

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/app/model"
	"BIP_backend/internal/service/auth"
	"BIP_backend/internal/service/mail"
)

var (
	incorrectUsernameOrPassword = errors.New("incorrect username or password")
	incorrectCode               = errors.New("incorrect code")
	incorrectToken              = errors.New("incorrect token")
	internalServerError         = errors.New("internal server error")
	tokenIsDeprecated           = errors.New("token is deprecated")
)

// @Summary      Registration
// @Description  registering a new account
// @Tags         api
// @Accept       json
// @Produce      json
// @Param        user_info   body  model.UserData  true  "info about user"
// @Success      200,400  {object}  structResponseUserCreate
// @Failure      500  {object}  errorResponse
// @Router       /registration [post]
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
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
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

type requestSessionsCreate struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// @Summary      Auth
// @Description  first step of two-factor authentication
// @Tags         api
// @Accept       json
// @Produce      json
// @Param        user_auth  body  requestSessionsCreate  true  "username and password"
// @Success      200 {object}  structResponseSessionsCreate
// @Failure      401,500  {object}  errorResponse
// @Router       /auth [post]
func (s *Server) handleSessionsCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var r = &requestSessionsCreate{}
		if err := c.ShouldBindJSON(r); err != nil {
			newErrorResponse(c, http.StatusUnauthorized, incorrectUsernameOrPassword)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}

		u, err := store.User().FindByUsername(r.Username)
		if err != nil || !u.ComparePassword(r.Password) {
			newErrorResponse(c, http.StatusUnauthorized, incorrectUsernameOrPassword)
			return
		}

		authorizer, err := auth.NewAuthorizer()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}
		jwt, err := authorizer.GenerateToken(u, false /* authorized */)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}
		code, err := authorizer.GeneratePassword()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}

		sender, err := mail.NewSender()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}

		// не знаю насколько тут уместно использовать горутину
		// с одной стороны ошибку надо обработать, но с другой нужно ответ отправить быстрее
		// (иначе оно с задержкой небольшой отправляет)
		// отпишись, что думаешь по этому поводу
		go sender.SendMail(u.Mail, code)

		cache, err := s.GetCache()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}
		if err := cache.Set(u.Username, code); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}

		c.JSON(http.StatusOK, responseSessionsCreate(jwt))
	}
}

type request2Factor struct {
	Code string `json:"code"`
}

// @Summary      Auth2Factor
// @Security 	 ApiKeyAuth
// @Description  second step of two-factor authentication
// @Tags         api
// @Accept       json
// @Produce      json
// @Param        code  body  request2Factor  true  "code sent by mail"
// @Success      200 {object}  structResponse2Factor
// @Failure      401,500  {object}  errorResponse
// @Router       /auth2fa [post]
func (s *Server) handler2Factor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var r = &request2Factor{}
		if err := c.ShouldBindJSON(r); err != nil {
			newErrorResponse(c, http.StatusUnauthorized, incorrectCode)
			return
		}

		username, ok := c.Get("username")
		if !ok {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken)
			return
		}

		cache, err := s.GetCache()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}
		code, err := cache.Get(username.(string))
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, tokenIsDeprecated)
			return
		}
		if code != r.Code {
			newErrorResponse(c, http.StatusUnauthorized, incorrectCode)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}

		u, err := store.User().FindByUsername(username.(string))
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken)
			return
		}

		authorizer, err := auth.NewAuthorizer()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
			return
		}
		jwt, err := authorizer.GenerateToken(u, true /* authorized */)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError)
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

func responseUserCreate(success bool) *structResponseUserCreate {
	return &structResponseUserCreate{
		Success: success,
	}
}

func responseSessionsCreate(jwt string) *structResponseSessionsCreate {
	return &structResponseSessionsCreate{
		JWT: jwt,
	}
}

func response2Factor(jwt string, user *model.User) *structResponse2Factor {
	return &structResponse2Factor{
		JWT:  jwt,
		User: user,
	}
}

func newErrorResponse(c *gin.Context, httpError int, definition error) {
	c.JSON(httpError, errorResponse{Error: definition.Error()})
}
