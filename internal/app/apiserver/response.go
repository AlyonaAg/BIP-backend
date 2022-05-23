package apiserver

import (
	"BIP_backend/internal/app/model"
	"BIP_backend/internal/app/store"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	Success bool `json:"success"`
}

type structBaseUserInfo struct {
	ID               int              `json:"id"`
	Rating           float64          `json:"rating"`
	Comment          []*model.Comment `json:"comment"`
	ListPhotoProfile []string         `json:"list_photo_profile"`
	Username         string           `json:"username"`
	FirstName        string           `json:"first_name"`
	SecondName       string           `json:"second_name"`
	AvatarURL        string           `json:"avatar_url"`
	PhoneNumber      string           `json:"phone_number"`
	IsPhotographer   bool             `json:"is_photographer"`
}

type structOrderForPhotographer struct {
	ID        int                 `json:"id"`
	Client    *structBaseUserInfo `json:"client"`
	OrderCost int                 `json:"order_cost"`
	// Location  model.Location      `json:"coordinates"`
	Comment string `json:"comment"`
	State   string `json:"state"`
}

type structBaseOrderInfoForPhotographer struct {
	OrderList []structOrderForPhotographer `json:"order_data"`
}

type structOrderForClient struct {
	ID           int                 `json:"id"`
	Photographer *structBaseUserInfo `json:"photographer"`
	OrderCost    int                 `json:"order_cost"`
	//Location     model.Location      `json:"coordinates"`
	Comment string `json:"comment"`
	State   string `json:"state"`
}

type structBaseOrderInfoForClient struct {
	OrderList []structOrderForClient `json:"order_data"`
}

type structResponseSessionsCreate struct {
	JWT string `json:"jwt"`
}

type structResponse2Factor struct {
	JWT  string              `json:"jwt"`
	User *structBaseUserInfo `json:"user"`
}

type structResponseAgreedPhotographers struct {
	Photographers []*structBaseUserInfo `json:"photographers"`
}

type structResponseAllPhotographers struct {
	Photographers []*structBaseUserInfo `json:"photographers"`
}

type structResponseFinishOrder struct {
	URLOriginal string `json:"url_original"`
}

type structResponseGetPreview struct {
	URLWatermark string `json:"url_watermark"`
}

type structResponseCreateQRCode struct {
	Code []byte `json:"code"`
}

type structResponseConfirmQRCode struct {
	Money int `json:"money"`
}

type structResponseGetMoney struct {
	Money int `json:"money"`
}

type structResponseGetOrdersForPhotographer struct {
	Backlog  []structOrderForPhotographer `json:"backlog"`
	Active   []structOrderForPhotographer `json:"active"`
	Finished []structOrderForPhotographer `json:"finished"`
}

type structResponseGetOrdersForClient struct {
	Backlog  []structOrderForClient `json:"backlog"`
	Active   []structOrderForClient `json:"active"`
	Finished []structOrderForClient `json:"finished"`
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

func responseGetAllPhotographer(photographers []model.User) *structResponseAllPhotographers {
	var photographersData = &structResponseAllPhotographers{}
	for _, u := range photographers {
		bu := getBaseUserInfo(&u)
		photographersData.Photographers = append(photographersData.Photographers, bu)
	}
	return photographersData
}

func responseFinishOrder(URLOrdinary string) *structResponseFinishOrder {
	return &structResponseFinishOrder{
		URLOriginal: URLOrdinary,
	}
}

func responseGetPreview(URLWatermark string) *structResponseGetPreview {
	return &structResponseGetPreview{
		URLWatermark: URLWatermark,
	}
}

func responseGetOrder(orders []model.Order, ur *store.UserRepository) (*structBaseOrderInfoForPhotographer, error) {
	ordersData, err := getBaseOrderInfoForPhotographer(orders, ur)
	if err != nil {
		return nil, err
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

func responseCreateQRCode(code []byte) *structResponseCreateQRCode {
	return &structResponseCreateQRCode{
		Code: code,
	}
}

func responseConfirmQRCode(money int) *structResponseConfirmQRCode {
	return &structResponseConfirmQRCode{
		Money: money,
	}
}

func responseGetMoney(money int) *structResponseGetMoney {
	return &structResponseGetMoney{
		Money: money,
	}
}

func responseGetOrdersForPhotographer(backlog, active, finished []model.Order,
	ur *store.UserRepository) (*structResponseGetOrdersForPhotographer, error) {
	backlogBaseInfo, err := getBaseOrderInfoForPhotographer(backlog, ur)
	if err != nil {
		return nil, err
	}

	activeBaseInfo, err := getBaseOrderInfoForPhotographer(active, ur)
	if err != nil {
		return nil, err
	}

	finishBaseInfo, err := getBaseOrderInfoForPhotographer(finished, ur)
	if err != nil {
		return nil, err
	}

	return &structResponseGetOrdersForPhotographer{
		Backlog:  backlogBaseInfo.OrderList,
		Active:   activeBaseInfo.OrderList,
		Finished: finishBaseInfo.OrderList,
	}, nil
}

func responseGetOrdersForClient(backlog, active, finished []model.Order,
	ur *store.UserRepository) (*structResponseGetOrdersForClient, error) {
	backlogBaseInfo, err := getBaseOrderInfoForClient(backlog, ur)
	if err != nil {
		return nil, err
	}

	activeBaseInfo, err := getBaseOrderInfoForClient(active, ur)
	if err != nil {
		return nil, err
	}

	finishBaseInfo, err := getBaseOrderInfoForClient(finished, ur)
	if err != nil {
		return nil, err
	}

	return &structResponseGetOrdersForClient{
		Backlog:  backlogBaseInfo.OrderList,
		Active:   activeBaseInfo.OrderList,
		Finished: finishBaseInfo.OrderList,
	}, nil
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
		IsPhotographer:   user.IsPhotographer,
	}
}

func getBaseOrderInfoForPhotographer(orders []model.Order,
	ur *store.UserRepository) (*structBaseOrderInfoForPhotographer, error) {
	var ordersData = &structBaseOrderInfoForPhotographer{}

	fmt.Println(orders)
	for _, order := range orders {
		u, err := ur.FindByID(order.ClientID)
		if err != nil {
			return nil, err
		}
		bu := getBaseUserInfo(u)
		var orderData = structOrderForPhotographer{
			ID:        order.ID,
			OrderCost: order.OrderCost,
			Client:    bu,
			//Location:  order.Location,
			Comment: order.Comment,
			State:   order.OrderState,
		}

		ordersData.OrderList = append(ordersData.OrderList, orderData)
	}
	return ordersData, nil
}

func getBaseOrderInfoForClient(orders []model.Order,
	ur *store.UserRepository) (*structBaseOrderInfoForClient, error) {
	var ordersData = &structBaseOrderInfoForClient{}
	fmt.Println(orders)

	for _, order := range orders {
		var bu *structBaseUserInfo
		if order.PhotographerID != 0 {
			u, err := ur.FindByID(order.PhotographerID)
			if err != nil {
				return nil, err
			}
			bu = getBaseUserInfo(u)
		}
		var orderData = structOrderForClient{
			ID:           order.ID,
			OrderCost:    order.OrderCost,
			Photographer: bu,
			//Location:     order.Location,
			Comment: order.Comment,
			State:   order.OrderState,
		}

		ordersData.OrderList = append(ordersData.OrderList, orderData)
	}
	return ordersData, nil
}
