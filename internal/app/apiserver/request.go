package apiserver

type structRequestSessionsCreate struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type structRequest2Factor struct {
	Code string `json:"code"`
}
