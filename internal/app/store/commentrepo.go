package store

import (
	"BIP_backend/internal/app/model"
	"errors"
)

type CommentRepository struct {
	store *Store
}

func (cr *CommentRepository) Create(orderID, userID, userComID, rating int, state, content string) error {
	store, err := cr.GetStore()
	if err != nil {
		return err
	}

	var ID int
	if err := store.db.QueryRow(
		`INSERT INTO "comments" (user_id, user_com_id, content, rating, state, order_id) `+
			`VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`, userID, userComID, content, rating, state, orderID,
	).Scan(
		&ID,
	); err != nil {
		return err
	}
	return nil
}

func (cr *CommentRepository) GetMeanRating(userID int) (float64, error) {
	store, err := cr.GetStore()
	if err != nil {
		return 0, err
	}

	var rating float64
	if err := store.db.QueryRow(
		`SELECT avg(rating) FROM "comments" WHERE user_id = $1`,
		userID,
	).Scan(&rating); err != nil {
		return 0, err
	}
	return rating, nil
}

func (cr *CommentRepository) GetListComment(userID int) ([]*model.Comment, error) {
	store, err := cr.GetStore()
	if err != nil {
		return nil, err
	}

	rows, err := store.db.Query(`SELECT user_com_id, content, rating `+
		`FROM "comments" WHERE user_id = $1`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment

	for rows.Next() {
		var comment = &model.Comment{}
		if err := rows.Scan(
			&comment.ClientID,
			&comment.Comment,
			&comment.Rating,
		); err != nil {
			continue
		}

		comment.ClientUsername, comment.AvatarURL, err = store.User().GetAvatarURLAndUsername(comment.ClientID)
		if err != nil {
			continue
		}

		comments = append(comments, comment)
	}
	return comments, nil
}

func (cr *CommentRepository) GetStore() (*Store, error) {
	if cr.store == nil {
		return nil, errors.New("empty comment store")
	}
	return cr.store, nil
}
