package apiserver

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"BIP_backend/internal/app/model"
	"BIP_backend/internal/app/store"
)

type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	Success bool `json:"success"`
}

type structBaseUserInfo struct {
	ID               int      `json:"id"`
	Rating           float64  `json:"rating"`
	Comment          []string `json:"comment"`
	ListPhotoProfile []string `json:"list_photo_profile"`
	Username         string   `json:"username"`
	FirstName        string   `json:"first_name"`
	SecondName       string   `json:"second_name"`
	AvatarURL        string   `json:"avatar_url"`
	PhoneNumber      string   `json:"phone_number"`
}

type structResponseSessionsCreate struct {
	JWT string `json:"jwt"`
}

type structResponse2Factor struct {
	JWT  string              `json:"jwt"`
	User *structBaseUserInfo `json:"user"`
}

type structOrder struct {
	ID        int                 `json:"id"`
	Client    *structBaseUserInfo `json:"client"`
	OrderCost int                 `json:"order_cost"`
	Location  model.Location      `json:"coordinates"`
}

type structResponseGetOrder struct {
	OrderList []structOrder `json:"order_data"`
}

type structResponseAgreedPhotographers struct {
	Photographers []*structBaseUserInfo `json:"photographers"`
}

func newSuccessResponse(success bool, err error) *successResponse {
	if err != nil {
		logrus.Error(err.Error())
	}
	return &successResponse{
		Success: success,
	}
}

func newErrorResponse(c *gin.Context, httpError int, definition error, msgLog error) {
	logrus.Error(msgLog.Error())
	c.JSON(httpError, errorResponse{Error: definition.Error()})
}

func responseSessionsCreate(jwt string) *structResponseSessionsCreate {
	return &structResponseSessionsCreate{
		JWT: jwt,
	}
}

func response2Factor(jwt string, user *model.User) *structResponse2Factor {
	return &structResponse2Factor{
		JWT:  jwt,
		User: getBaseUserInfo(user),
	}
}

func responseGetOrder(orders []model.Order, ur *store.UserRepository) (*structResponseGetOrder, error) {
	var ordersData = &structResponseGetOrder{}

	for _, order := range orders {
		u, err := ur.FindByID(order.ClientID)
		if err != nil {
			return nil, err
		}
		bu := getBaseUserInfo(u)
		var orderData = structOrder{
			ID:        order.ID,
			OrderCost: order.OrderCost,
			Client:    bu,
			Location:  order.Location,
		}

		ordersData.OrderList = append(ordersData.OrderList, orderData)
	}
	return ordersData, nil
}

func responseGetAgreedPhotographer(photographerID []int, ur *store.UserRepository) (
	*structResponseAgreedPhotographers, error) {
	var photographersData = &structResponseAgreedPhotographers{}

	for _, id := range photographerID {
		u, err := ur.FindByID(id)
		if err != nil {
			return nil, err
		}
		bu := getBaseUserInfo(u)

		photographersData.Photographers = append(photographersData.Photographers, bu)
	}
	return photographersData, nil
}

func getBaseUserInfo(user *model.User) *structBaseUserInfo {
	return &structBaseUserInfo{
		ID:               user.ID,
		Username:         user.Username,
		FirstName:        user.FirstName,
		SecondName:       user.SecondName,
		PhoneNumber:      user.PhoneNumber,
		AvatarURL:        user.AvatarURL,
		Rating:           user.Rating,
		Comment:          user.Comment,
		ListPhotoProfile: user.ListPhotoProfile,
	}
}
