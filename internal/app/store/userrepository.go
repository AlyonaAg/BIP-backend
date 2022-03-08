package store

import (
	"BIP_backend/internal/app/model"
	"errors"
)

type UserRepository struct {
	store *Store
}

func (ur *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	store, err := ur.GetStore()
	if err != nil {
		return err
	}

	if err := store.db.QueryRow(
		`INSERT INTO "user" (username, password, first_name, second_name,`+
			`is_photographer, avatar_url, phone_number, mail) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		u.Username, u.Password, u.FirstName, u.SecondName, u.IsPhotographer, u.AvatarURL,
		u.PhoneNumber, u.Mail,
	).Scan(
		&u.ID,
	); err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) FindByUsername(username string) (*model.User, error) {
	store, err := ur.GetStore()
	if err != nil {
		return nil, err
	}
	// add validation username

	var u = &model.User{}
	if err := store.db.QueryRow(
		`SELECT id, username, password, first_name, second_name, is_photographer, `+
			`avatar_url, phone_number, mail FROM "user" WHERE username = $1`,
		username,
	).Scan(
		&u.ID,
		&u.Username,
		&u.Password,
		&u.FirstName,
		&u.SecondName,
		&u.IsPhotographer,
		&u.AvatarURL,
		&u.PhoneNumber,
		&u.Mail,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (ur *UserRepository) GetStore() (*Store, error) {
	if ur.store == nil {
		return nil, errors.New("empty user store")
	}
	return ur.store, nil
}
