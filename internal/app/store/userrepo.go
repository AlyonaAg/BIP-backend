package store

import (
	"errors"

	"BIP_backend/internal/app/model"
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

	baseRating := 5
	baseMoney := 1000
	if err := store.db.QueryRow(
		`INSERT INTO "user" (username, password, first_name, second_name,`+
			`is_photographer, avatar_url, phone_number, mail, money, rating) VALUES ($1, $2, $3,`+
			`$4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		u.Username, u.Password, u.FirstName, u.SecondName, u.IsPhotographer, u.AvatarURL,
		u.PhoneNumber, u.Mail, baseMoney, baseRating,
	).Scan(
		&u.ID,
	); err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) FindByID(id int) (*model.User, error) {
	store, err := ur.GetStore()
	if err != nil {
		return nil, err
	}

	var u = &model.User{}
	if err := store.db.QueryRow(
		`SELECT id, username, password, first_name, second_name, is_photographer, `+
			`avatar_url, phone_number, mail, money, rating FROM "user" WHERE id = $1`,
		id,
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
		&u.Money,
		&u.Rating,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (ur *UserRepository) FindByUsername(username string) (*model.User, error) {
	store, err := ur.GetStore()
	if err != nil {
		return nil, err
	}

	var u = &model.User{}
	if err := store.db.QueryRow(
		`SELECT id, username, password, first_name, second_name, is_photographer, `+
			`avatar_url, phone_number, mail, money, rating FROM "user" WHERE username = $1`,
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
		&u.Money,
		&u.Rating,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (ur *UserRepository) WithdrawMoney(username string, money int) error {
	store, err := ur.GetStore()
	if err != nil {
		return err
	}

	if _, err := store.db.Exec(
		`UPDATE "user" SET money = money - $1 WHERE username = $2`, money, username); err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) PutMoney(username string, money int) error {
	store, err := ur.GetStore()
	if err != nil {
		return err
	}

	if _, err := store.db.Exec(
		`UPDATE "user" SET money = money + $1 WHERE username = $2`, money, username); err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) GetStore() (*Store, error) {
	if ur.store == nil {
		return nil, errors.New("empty user store")
	}
	return ur.store, nil
}
