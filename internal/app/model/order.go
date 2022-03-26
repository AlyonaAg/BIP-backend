package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

const (
	Created            string = "created"
	AgreedPhotographer        = "agreed_photographer"
	AgreedClient              = "agreed_client"
	Meeting                   = "meeting"
	WatermarkSent             = "watermarks_sent"
	Finish                    = "finish"
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type OrderData struct {
	Location  Location `json:"coordinates"`
	ClientID  int      `json:"client_id"`
	OrderCost int      `json:"order_cost"`
	Comment   string   `json:"comment"`
}

type Order struct {
	OrderData
	ID                int      `json:"id"`
	PhotographerID    int      `json:"photographer_id"`
	ClientCurLocation Location `json:"client_current_location"`
	OrderState        string   `json:"order_state"`
}

func (o *Order) Validate() error {
	return validation.ValidateStruct(
		o,
		validation.Field(&o.OrderCost, validation.Required, validation.Min(100)),
		validation.Field(&o.ClientID, validation.Required),
		validation.Field(&o.Location, validation.Required),
	)
}
