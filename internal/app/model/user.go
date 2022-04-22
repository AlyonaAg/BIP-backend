package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

type UserData struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	FirstName      string `json:"first_name"`
	SecondName     string `json:"second_name"`
	IsPhotographer bool   `json:"is_photographer"`
	AvatarURL      string `json:"avatar_url"`
	PhoneNumber    string `json:"phone_number"`
	Mail           string `json:"mail"`
}

type User struct {
	UserData
	ID               int        `json:"id"`
	Money            int        `json:"money"`
	Rating           float64    `json:"rating"`
	Comment          []*Comment `json:"comment"`
	ListPhotoProfile []string   `json:"list_photo_profile"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Username, validation.Required, validation.Length(2, 30), is.Alphanumeric),
		validation.Field(&u.Password, validation.Required, validation.Length(5, 100)),
		validation.Field(&u.FirstName, validation.Required, validation.Length(1, 15), is.Alpha),
		validation.Field(&u.SecondName, validation.Required, validation.Length(1, 15), is.Alpha),
		validation.Field(&u.IsPhotographer),
		validation.Field(&u.AvatarURL, is.URL),
		validation.Field(&u.PhoneNumber, is.E164),
		validation.Field(&u.Mail, validation.Required, is.Email),
	)
}

func (u *User) BeforeCreate() error {
	enc, err := encryptString(u.Password)
	if err != nil {
		return nil
	}
	u.Password = enc

	if u.AvatarURL == "" {
		u.AvatarURL = "https://yt3.ggpht.com/ytc/AKedOLSio27bVkyp7ExnEPfVHealr73MJkXXH0VlSi_f=s900-c-k-c0x00ffffff-no-rj"
	}
	return nil
}

func (u *User) ComparePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func encryptString(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
