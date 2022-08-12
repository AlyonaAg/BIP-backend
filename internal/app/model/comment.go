package model

type Comment struct {
	ClientID       int     `json:"client_id"`
	ClientUsername string  `json:"username"`
	AvatarURL      string  `json:"avatar_url"`
	Comment        string  `json:"content"`
	Rating         float64 `json:"rating"`
}
