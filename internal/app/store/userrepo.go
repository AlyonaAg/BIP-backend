package store

import (
	"BIP_backend/internal/app/model"
	"errors"
)

type UserRepository struct {
	store *Store
}

func (ur *UserRepository) Create(u *model.User, key string) error {
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
	var baseMoney = 0
	if !u.IsPhotographer {
		baseMoney = 1000
	}
	if err := store.db.QueryRow(
		`INSERT INTO "user" (username, password, first_name, second_name,`+
			`is_photographer, avatar_url, phone_number, mail, money, rating, secret_key) VALUES ($1, $2, $3,`+
			`$4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		u.Username, u.Password, u.FirstName, u.SecondName, u.IsPhotographer, u.AvatarURL,
		u.PhoneNumber, u.Mail, baseMoney, baseRating, key,
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

	u.Comment, err = store.Comment().GetListComment(u.ID)
	if err != nil {
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

	u.Comment, err = store.Comment().GetListComment(u.ID)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ur *UserRepository) GetAllPhotographer(page int) ([]model.User, error) {
	store, err := ur.GetStore()
	if err != nil {
		return nil, err
	}

	var offset = (page - 1) * 10

	rows, err := store.db.Query(`SELECT id, username, password, first_name, second_name, is_photographer, `+
		`avatar_url, phone_number, mail, money, rating FROM "user" WHERE is_photographer = true LIMIT 10 OFFSET $1`,
		offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listPhotographers []model.User
	for rows.Next() {
		var u = model.User{}
		if err := rows.Scan(
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
			continue
		}

		u.Comment, err = store.Comment().GetListComment(u.ID)
		if err != nil {
			continue
		}

		listPhotographers = append(listPhotographers, u)
	}
	return listPhotographers, nil
}

func (ur *UserRepository) GetAvatarURLAndUsername(userID int) (string /*username*/, string /*avatar_url*/, error) {
	store, err := ur.GetStore()
	if err != nil {
		return "", "", err
	}

	var username, avatarURL string
	if err := store.db.QueryRow(
		`SELECT username, avatar_url FROM "user" WHERE id = $1`,
		userID,
	).Scan(&username, &avatarURL); err != nil {
		return "", "", err
	}

	return username, avatarURL, nil
}

func (ur *UserRepository) GetMoney(userID int) (int, error) {
	store, err := ur.GetStore()
	if err != nil {
		return 0, err
	}

	var money int
	if err := store.db.QueryRow(
		`SELECT money FROM "user" WHERE id = $1`,
		userID,
	).Scan(&money); err != nil {
		return 0, err
	}

	return money, nil
}

func (ur *UserRepository) WithdrawMoneyByID(userID string, money int) error {
	store, err := ur.GetStore()
	if err != nil {
		return err
	}

	if _, err := store.db.Exec(
		`UPDATE "user" SET money = money - $1 WHERE id = $2`, money, userID); err != nil {
		return err
	}
	return nil
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

func (ur *UserRepository) PutMoneyByID(userID, money int) error {
	store, err := ur.GetStore()
	if err != nil {
		return err
	}

	if _, err := store.db.Exec(
		`UPDATE "user" SET money = money + $1 WHERE id = $2`, money, userID); err != nil {
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

func (ur *UserRepository) UpdateRating(userID int, rating float64) error {
	store, err := ur.GetStore()
	if err != nil {
		return err
	}

	if _, err := store.db.Exec(
		`UPDATE "user" SET rating = $1 WHERE id = $2`, rating, userID); err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) GetSecretKey(userID int) (string, error) {
	store, err := ur.GetStore()
	if err != nil {
		return "", err
	}

	var key string
	if err := store.db.QueryRow(
		`SELECT secret_key FROM "user" WHERE id = $1`,
		userID,
	).Scan(&key); err != nil {
		return "", err
	}

	return key, nil
}

func (ur *UserRepository) CheckSecretKey(userID int, secretKey string) error {
	store, err := ur.GetStore()
	if err != nil {
		return err
	}

	var ID int
	if err := store.db.QueryRow(
		`SELECT id FROM "user" WHERE id = $1 AND secret_key = $2`,
		userID, secretKey,
	).Scan(&ID); err != nil {
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
