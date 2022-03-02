package model

type User struct {
	First_name      string `json:"first_name"`
	Second_name     string `json:"second_name"`
	Is_photographer bool   `json:"is_photographer"`
	Avatar_URL      string `json:"avatar_url"`
	Phone_number    string `json:"phone_number"`
	Mail            string `json:"mail"`
}
