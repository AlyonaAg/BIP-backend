package apiserver

type structRequestSessionsCreate struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type structRequest2Factor struct {
	Code string `json:"code"`
}

type structRequestUpload struct {
	URLOriginal  string `json:"url_origin"`
	URLWatermark string `json:"url_watermark"`
}

type structRequestReview struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}