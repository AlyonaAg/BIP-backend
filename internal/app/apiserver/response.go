package apiserver

import "BIP_backend/internal/app/model"

type errorResponse struct {
	Error string `json:"error"`
}

type structResponseUserCreate struct {
	Success bool `json:"success"`
}

type structResponseSessionsCreate struct {
	JWT string `json:"jwt"`
}

type structResponse2Factor struct {
	JWT  string      `json:"jwt"`
	User *model.User `json:"user"`
}
