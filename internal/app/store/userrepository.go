package store

import "BIP_backend/internal/app/model"

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	if err := r.store.db.QueryRow(
		`INSERT INTO "user" (username, password, first_name, second_name,`+
			`is_photographer, avatar_url, phone_number, mail) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		u.Username, u.Password, u.FirstName, u.SecondName, u.IsPhotographer, u.AvatarURL,
		u.PhoneNumber, u.Mail).Scan(&u.ID); err != nil {
		return err
	}
	return nil
}

func (s *UserRepository) FindByUsername(username string, password string) (*model.User, error) {
	return nil, nil
}
