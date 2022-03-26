package apiserver

import (
	"BIP_backend/internal/service/crypt"
	"BIP_backend/internal/service/qrcode"
	"encoding/hex"
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
	incorrectClientID           = errors.New("incorrect user id")
	internalServerError         = errors.New("internal server error")
	insufficientFunds           = errors.New("insufficient funds")
	tokenIsDeprecated           = errors.New("token is deprecated")
	incorrectOrderID            = errors.New("incorrect order id")
	incorrectPhotographerID     = errors.New("incorrect photographer id")
	incorrectAccept             = errors.New("incorrect accept")
	incorrectLocation           = errors.New("incorrect location")
	incorrectQRCode             = errors.New("incorrect QR-code")
	orderCompleted              = errors.New("order completed")
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
		key, err := crypt.GenerateRandKey(16)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.User().Create(u, hex.EncodeToString(key)); err != nil {
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
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, incorrectUsernameOrPassword, err)
			return
		}
		if !u.ComparePassword(r.Password) {
			newErrorResponse(c, http.StatusUnauthorized, incorrectUsernameOrPassword, incorrectUsernameOrPassword)
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

		cache, err := s.GetPassCache()
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

		cache, err := s.GetPassCache()
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

		c.JSON(http.StatusOK, newSuccessResponse(true, nil))
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
// @Router       /client/offer [get]
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

// @Summary      Get list all photographers
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Success      200  {object}  structResponseAllPhotographers
// @Failure      500  {object}  errorResponse
// @Router       /client/photographers [get]
func (s *Server) handlerGetAllPhotographer() gin.HandlerFunc {
	return func(c *gin.Context) {
		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		photographers, err := store.User().GetAllPhotographer()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		c.JSON(http.StatusOK, responseGetAllPhotographer(photographers))
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

// @Summary      Finish order
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Param        id_order  query  int  true  "id order"
// @Success      200  {object}  structResponseFinishOrder
// @Failure      500  {object}  errorResponse
// @Router       /client/finish-order [POST]
func (s *Server) handlerFinishOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.Order().UpdateOrderState(model.Finish, orderID); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		URLOriginal, err := store.Order().GetURLOriginal(orderID)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, err, err)
			return
		}

		cost, err := store.Order().GetCost(orderID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		photographerID, err := store.Order().GetPhotographerID(orderID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		store.User().PutMoneyByID(photographerID, int(float64(cost)-float64(cost)*0.3))

		c.JSON(http.StatusOK, responseFinishOrder(URLOriginal))
	}
}

// @Summary      Upload link
// @Security 	 ApiKeyAuth
// @Tags         photographer api
// @Accept       json
// @Produce      json
// @Param        id_order  query  int  true  "id order"
// @Param        link  body  structRequestUpload  true  "link"
// @Success      200  {object}  successResponse
// @Failure      400  {object}  successResponse
// @Failure      500  {object}  errorResponse
// @Router       /ph/upload [post]
func (s *Server) handlerUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
			return
		}

		var r = &structRequestUpload{}
		if err := c.ShouldBindJSON(r); err != nil {
			c.JSON(http.StatusBadRequest, newSuccessResponse(false, err))
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.Order().UpdateURL(r.URLOriginal, r.URLWatermark, orderID); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		c.JSON(http.StatusOK, newSuccessResponse(true, nil))
	}
}

// @Summary      Get preview
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Param        id_order  query  int  true  "id order"
// @Success      200  {object}  structResponseGetPreview
// @Failure      500  {object}  errorResponse
// @Router       /client/get-preview [GET]
func (s *Server) handlerGetPreview() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.Order().UpdateOrderState(model.WatermarkSent, orderID); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		URLWatermark, err := store.Order().GetURLWatermark(orderID)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, err, err)
			return
		}

		c.JSON(http.StatusOK, responseGetPreview(URLWatermark))
	}
}

// @Summary      Get QR-code
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Param        id_order  query  int  true  "id order"
// @Param        id_user   query  int  true  "id user"
// @Param        latitude  query  float64  true  "latitude"
// @Param        longitude  query  float64  true  "longitude"
// @Success      200  {object}  structResponseCreateQRCode
// @Failure      400,500  {object}  errorResponse
// @Router       /client/qrcode [GET]
func (s *Server) handlerCreateQRCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			return
		}
		userID, err := strconv.Atoi(c.Query("id_user"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}
		latitude, err := strconv.ParseFloat(c.Query("latitude"), 64)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectLocation, err)
			return
		}
		longitude, err := strconv.ParseFloat(c.Query("longitude"), 64)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectLocation, err)
			return
		}

		var location = &model.Location{
			Latitude:  latitude,
			Longitude: longitude,
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.Order().UpdateCurrentLocation(location, orderID); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		secretKey, err := store.User().GetSecretKey(userID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		tempKey, err := crypt.GenerateRandKey(32)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		cache, err := s.GetKeyCache()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		if err := cache.Set(c.Query("id_order"), hex.EncodeToString(tempKey)); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		secret, err := crypt.EncryptAES(tempKey, secretKey)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		qrCoder, err := qrcode.NewQRCoder()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		code, err := qrCoder.CreateQRCode(location, orderID, secret)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		c.JSON(http.StatusOK, responseCreateQRCode(code))
	}
}

// @Summary      Confirm QR-code
// @Security 	 ApiKeyAuth
// @Tags         photographer api
// @Accept       json
// @Produce      json
// @Param        qrcode  query  string  true  "qr-code"
// @Success      200  {object}  structResponseCreateQRCode
// @Failure      400,500  {object}  errorResponse
// @Router       /ph/confirm-qrcode [PATCH]
func (s *Server) handlerConfirmQRCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		qrCode := c.Query("qrcode")
		qrCoder, err := qrcode.NewQRCoder()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		photographerID, ok := c.Get("user_id")
		if !ok {
			newErrorResponse(c, http.StatusBadRequest, incorrectToken, incorrectToken)
			return
		}

		QRLocation, QROrderID, QRSecret, err := qrCoder.DecodeQRCode(qrCode)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectQRCode, err)
			return
		}

		cache, err := s.GetKeyCache()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		tempKey, err := cache.Get(strconv.Itoa(QROrderID))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectQRCode, err)
			return
		}

		tempKeyByte, err := hex.DecodeString(tempKey)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectQRCode, err)
			return
		}
		secretKey, err := crypt.DecryptAES(tempKeyByte, QRSecret)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectQRCode, err)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		clientID, orderCost, err := store.Order().GetClientIDByQRInfo(QRLocation, QROrderID, photographerID.(int))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectQRCode, err)
			return
		}
		if err := store.User().CheckSecretKey(clientID, secretKey); err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectQRCode, err)
			return
		}
		if err := store.Order().UpdateOrderState(model.Meeting, QROrderID); err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectQRCode, err)
			return
		}

		store.User().PutMoneyByID(photographerID.(int), int(float64(orderCost)*0.3))

		c.JSON(http.StatusOK, responseConfirmQRCode(int(float64(orderCost)*0.3)))
	}
}

// @Summary      Client feedback
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Param        id_order  query  string  true  "id order"
// @Param        id_client  query  string  true  "id client"
// @Param        review  body  structRequestReview  true  "review"
// @Success      200  {object}  successResponse
// @Failure      400,500  {object}  errorResponse
// @Router       /client/review [POST]
func (s *Server) handlerClientReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			return
		}
		clientID, err := strconv.Atoi(c.Query("id_client"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}

		var r = &structRequestReview{}
		if err := c.ShouldBindJSON(r); err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectRequestData, err)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		order, err := store.Order().GetOrderByID(orderID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.Comment().Create(order.PhotographerID, clientID, r.Rating, model.Finish, r.Comment); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		rating, err := store.Comment().GetMeanRating(order.PhotographerID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		if err := store.User().UpdateRating(order.PhotographerID, rating); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		c.JSON(http.StatusOK, newSuccessResponse(true, nil))
	}
}

// @Summary      Photographer feedback
// @Security 	 ApiKeyAuth
// @Tags         photographer api
// @Accept       json
// @Produce      json
// @Param        id_order  query  string  true  "id order"
// @Param        id_user  query  string  true  "id photographer"
// @Param        review  body  structRequestReview  true  "review"
// @Success      200  {object}  successResponse
// @Failure      400,500  {object}  errorResponse
// @Router       /ph/review [POST]
func (s *Server) handlerPhotographerReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			return
		}
		photographID, err := strconv.Atoi(c.Query("id_user"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}

		var r = &structRequestReview{}
		if err := c.ShouldBindJSON(r); err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectRequestData, err)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		order, err := store.Order().GetOrderByID(orderID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if err := store.Comment().Create(order.ClientID, photographID, r.Rating, model.Finish, r.Comment); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		rating, err := store.Comment().GetMeanRating(order.ClientID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		if err := store.User().UpdateRating(order.ClientID, rating); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		c.JSON(http.StatusOK, newSuccessResponse(true, nil))
	}
}

// @Summary      Order cancellation
// @Security 	 ApiKeyAuth
// @Tags         client api
// @Accept       json
// @Produce      json
// @Param        id_order  query  string  true  "id order"
// @Param        id_user  query  string  true  "id user"
// @Param        review  body  structRequestReview  true  "review"
// @Success      200  {object}  successResponse
// @Failure      400,500  {object}  errorResponse
// @Router       /client/cancel [POST]
func (s *Server) handlerCancel() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			return
		}
		clientID, err := strconv.Atoi(c.Query("id_user"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}

		var r = &structRequestReview{}
		if err := c.ShouldBindJSON(r); err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectRequestData, err)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		order, err := store.Order().GetOrderByID(orderID)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}

		if order.OrderState == model.Finish {
			newErrorResponse(c, http.StatusBadRequest, orderCompleted, orderCompleted)
			return
		}

		if order.PhotographerID != 0 {
			if err := store.Comment().Create(order.PhotographerID, clientID, r.Rating, order.OrderState, r.Comment); err != nil {
				newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
				return
			}
			rating, err := store.Comment().GetMeanRating(order.PhotographerID)
			if err != nil {
				newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
				return
			}
			if err := store.User().UpdateRating(order.PhotographerID, rating); err != nil {
				newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
				return
			}
		}

		store.User().PutMoneyByID(clientID, int(float64(order.OrderCost)-float64(order.OrderCost)*0.3))

		if err := store.Order().UpdateOrderState(model.Finish, orderID); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		c.JSON(http.StatusOK, newSuccessResponse(true, nil))
	}
}

// @Summary      Info about user
// @Security 	 ApiKeyAuth
// @Tags         api
// @Accept       json
// @Produce      json
// @Param        id_user  query  string  true  "id user"
// @Success      200  {object}  structBaseUserInfo
// @Failure      400,500  {object}  errorResponse
// @Router       /profile [GET]
func (s *Server) handlerProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Query("id_user"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		user, err := store.User().FindByID(userID)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}

		c.JSON(http.StatusOK, getBaseUserInfo(user))
	}
}

// @Summary      User money data
// @Security 	 ApiKeyAuth
// @Tags         api
// @Accept       json
// @Produce      json
// @Param        id_user  query  string  true  "id user"
// @Success      200  {object}  structResponseGetMoney
// @Failure      400,500  {object}  errorResponse
// @Router       /get-money [GET]
func (s *Server) handlerGetMoney() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Query("id_user"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}

		tokenUserID, ok := c.Get("user_id")
		if !ok {
			newErrorResponse(c, http.StatusUnauthorized, incorrectToken, incorrectToken)
			return
		}

		if tokenUserID != userID {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, incorrectClientID)
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		money, err := store.User().GetMoney(userID)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			return
		}

		c.JSON(http.StatusOK, responseGetMoney(money))
	}
}

func (s *Server) checkOrderForClient() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			c.Abort()
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			c.Abort()
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		if err := store.Order().CheckMatchingClientAndOrderID(orderID, userID.(int)); err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			c.Abort()
			return
		}
	}
}

func (s *Server) checkOrderForPhotographer() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.Atoi(c.Query("id_order"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			c.Abort()
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			newErrorResponse(c, http.StatusBadRequest, incorrectClientID, err)
			c.Abort()
			return
		}

		store, err := s.GetStore()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, internalServerError, err)
			return
		}
		if err := store.Order().CheckMatchingPhotographAndOrderID(orderID, userID.(int)); err != nil {
			newErrorResponse(c, http.StatusBadRequest, incorrectOrderID, err)
			c.Abort()
			return
		}
	}
}
