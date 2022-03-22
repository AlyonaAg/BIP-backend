package apiserver

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/app/model"
	"BIP_backend/internal/service/auth"
	"BIP_backend/internal/service/mail"
)

var (
	incorrectRequestData        = errors.New("incorrect request data")
	incorrectUsernameOrPassword = errors.New("incorrect username or password")
	incorrectCode               = errors.New("incorrect code")
	incorrectToken              = errors.New("incorrect token")
	incorrectClientID           = errors.New("id from request does not match username from token")
	internalServerError         = errors.New("internal server error")
	insufficientFunds           = errors.New("insufficient funds")
	tokenIsDeprecated           = errors.New("token is deprecated")
	incorrectOrderID            = errors.New("incorrect order id")
	incorrectPhotographerID     = errors.New("incorrect photographer id")
	incorrectAccept             = errors.New("incorrect accept")
)

// @Summary      Registration
// @Description  registering a new account
// @Tags         api
// @Accept       json
// @Produce      json
// @Param        user_info   body  model.UserData  true  "info about user"
// @Success      200,400  {object}  successResponse
// @Failure      500  {object}  errorResponse
// @Router       /registration [post]
func (s *Server) handleUserCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var u = &model.User{}
		if err := c.ShouldBindJSON(u); err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.User().Create(u); err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
			return
		}
		c.JSON(http.StatusOK, newSuccessResponse(true, nil))
	}
}

// @Summary      Auth
// @Description  first step of two-factor authentication
// @Tags         api
// @Accept       json
// @Produce      json
// @Param        user_auth  body  structRequestSessionsCreate  true  "username and password"
// @Success      200 {object}  structResponseSessionsCreate
// @Failure      401,500  {object}  errorResponse
// @Router       /auth [post]
func (s *Server) handleSessionsCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var r = &structRequestSessionsCreate{}
		if err := c.ShouldBindJSON(r); err != nil {
			newErrorResponse(c, http.StatusUnauthorized, incorrectUsernameOrPassword, err)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		u, err := store.User().FindByUsername(r.Username)
		if err != nil || !u.ComparePassword(r.Password) {
			newErrorResponse(c, http.StatusUnauthorized, incorrectUsernameOrPassword, err)
			return
		}

		authorizer, err := auth.NewAuthorizer()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		jwt, err := authorizer.GenerateToken(u, false /* authorized */)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		code, err := authorizer.GeneratePassword()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		sender, err := mail.NewSender()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		// не знаю насколько тут уместно использовать горутину
		// с одной стороны ошибку надо обработать, но с другой нужно ответ отправить быстрее
		// (иначе оно с задержкой небольшой отправляет)
		// отпишись, что думаешь по этому поводу
		go sender.SendMail(u.Mail, code)

		cache, err := s.GetCache()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		if err := cache.Set(strconv.Itoa(u.ID), code); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		c.JSON(http.StatusOK, responseSessionsCreate(jwt))
	}
}

// @Summary      Auth2Factor
// @Security 	 ApiKeyAuth
// @Description  second step of two-factor authentication
// @Tags         api
// @Accept       json
// @Produce      json
// @Param        code  body  structRequest2Factor  true  "code sent by mail"
// @Success      200 {object}  structResponse2Factor
// @Failure      401,500  {object}  errorResponse
// @Router       /auth2fa [post]
func (s *Server) handler2Factor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var r = &structRequest2Factor{}
		if err := c.ShouldBindJSON(r); err != nil {
			newErrorResponse(c, http.StatusUnauthorized, incorrectCode, err)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken, incorrectToken)
			return
		}

		cache, err := s.GetCache()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		code, err := cache.Get(strconv.Itoa(userID.(int)))
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, tokenIsDeprecated, err)
			return
		}
		if code != r.Code {
			newErrorResponse(c, http.StatusUnauthorized, incorrectCode, incorrectCode)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		u, err := store.User().FindByID(userID.(int))
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken, err)
			return
		}

		authorizer, err := auth.NewAuthorizer()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		jwt, err := authorizer.GenerateToken(u, true /*authorized*/)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		cache.Del(strconv.Itoa(userID.(int)))
		c.JSON(http.StatusOK, response2Factor(jwt, u))
	}
}

// @Summary      Create order
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Param        order  body  model.OrderData  true  "order data"
// @Success      200 {object}  model.Order
// @Failure      500,400  {object}  errorResponse
// @Router       /client/create-order [post]
func (s *Server) handlerCreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var o = &model.Order{}
		if err := c.ShouldBindJSON(o); err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectRequestData, err)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			newErrorResponse(c, http.StatusBadRequest, incorrectToken, incorrectToken)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if userID.(int) != o.ClientID {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, incorrectClientID)
			return
		}

		u, err := store.User().FindByID(o.ClientID)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}
		if u.Money < o.OrderCost {
			newErrorResponse(c, http.StatusBadRequest, insufficientFunds, insufficientFunds)
			return
		}

		o.OrderState = model.Created
		if err := store.User().WithdrawMoney(u.Username, o.OrderCost); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.Order().Create(o); err != nil {
			store.User().PutMoney(u.Username, o.OrderCost)
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		c.JSON(http.StatusOK, o)
	}
}

// @Summary      Get order list
// @Security 	 ApiKeyAuth
// @Tags         photographer api
// @Accept       json
// @Produce      json
// @Success      200 {object}  structResponseGetOrder
// @Failure      500  {object}  errorResponse
// @Router       /ph/orders [get]
func (s *Server) handlerGetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		orders, err := store.Order().GetListCreatedOrder()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		response, err := responseGetOrder(orders, store.User())
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

// @Summary      Select order
// @Security 	 ApiKeyAuth
// @Description  The photographer chooses which orders he is ready to accept
// @Tags         photographer api
// @Accept       json
// @Produce      json
// @Param        id_order  query  int  true  "id order"
// @Param        id_photographer query  int  true  "id photographer"
// @Success      200,400 {object}  successResponse
// @Failure      500  {object}  errorResponse
// @Router       /ph/select [patch]
func (s *Server) handlerSelect() gin.HandlerFunc {
	return func(c *gin.Context) {
		idOrder, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, incorrectOrderID))
			return
		}
		idPhotographer, err := strconv.Atoi(c.Query("id_photographer"))
		if err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, incorrectPhotographerID))
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		if err := store.Photographer().Create(idOrder, idPhotographer); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.Order().UpdateOrderState(model.AgreedPhotographer, idOrder); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		c.JSON(http.StatusBadRequest, newSuccessResponse(true, nil))
	}
}

// @Summary      Get list agreed photographers
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Param        id_order  query  int  true  "id order"
// @Success      200  {object}  structResponseAgreedPhotographers
// @Failure      400  {object}  successResponse
// @Failure      500  {object}  errorResponse
// @Router       /client/photographers [get]
func (s *Server) handlerGetAgreedPhotographer() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, incorrectOrderID))
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		photographersID, err := store.Photographer().GetListPhotographerByOrderID(orderID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		response, err := responseGetAgreedPhotographer(photographersID, store.User())
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

// @Summary      Accept photographer
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Param        id_order  query  int  true  "id order"
// @Param        id_photographer  query  int  true  "id order"
// @Param        is_accept  query  bool  true  "id order"
// @Success      200  {object}  successResponse
// @Failure      400  {object}  successResponse
// @Failure      500  {object}  errorResponse
// @Router       /client/accept [patch]
func (s *Server) handlerAccept() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, incorrectOrderID))
			return
		}
		photographerID, err := strconv.Atoi(c.Query("id_photographer"))
		if err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, incorrectPhotographerID))
			return
		}
		isAccept, err := strconv.ParseBool(c.Query("is_accept"))
		if err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, incorrectAccept))
			return
		}

		store, err := s.GetStore()
		if err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
			return
		}

		if err := store.Photographer().CheckOrderAvailability(photographerID, orderID); err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
			return
		}
		if isAccept {
			if err := store.Order().UpdateOrderState(model.AgreedClient, orderID); err != nil {
				c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
				return
			}
			if err := store.Order().UpdateOrderPhotographer(photographerID, orderID); err != nil {
				c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
				return
			}
			if err := store.Photographer().DelAllByOrderID(orderID); err != nil {
				c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
				return
			}
		} else {
			if err := store.Photographer().DelPhotographerByOrderID(photographerID, orderID); err != nil {
				c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
				return
			}
		}

		c.JSON(http.StatusOK, newSuccessResponse(true, nil))
	}
}
