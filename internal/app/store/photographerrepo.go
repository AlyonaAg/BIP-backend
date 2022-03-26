package store

import (
	"errors"
)

type PhotographerRepository struct {
	store *Store
}

func (phr *PhotographerRepository) Create(orderID, userID int) error {
	store, err := phr.GetStore()
	if err != nil {
		return err
	}

	var ID int
	if err := store.db.QueryRow(
		`INSERT INTO "agreed_photographers" (order_id, photographer_id) `+
			`VALUES ($1, $2) RETURNING id`, orderID, userID,
	).Scan(
		&ID,
	); err != nil {
		return err
	}
	return nil
}

func (phr *PhotographerRepository) GetListPhotographerByOrderID(orderID int) ([]int, error) {
	store, err := phr.GetStore()
	if err != nil {
		return nil, err
	}

	rows, err := store.db.Query(`SELECT photographer_id `+
		`FROM "agreed_photographers" WHERE order_id = $1`, orderID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listID []int

	for rows.Next() {
		var ID int
		if err := rows.Scan(&ID); err != nil {
			return listID, err
		}
		listID = append(listID, ID)
	}
	return listID, err
}

func (phr *PhotographerRepository) CheckOrderAvailability(photographerID, orderID int) error {
	store, err := phr.GetStore()
	if err != nil {
		return err
	}

	rows, err := store.db.Exec(
		`SELECT * FROM "agreed_photographers" WHERE photographer_id = $1 AND order_id = $2`,
		photographerID, orderID)
	if err != nil {
		return err
	}
	if count, _ := rows.RowsAffected(); count == 0 {
		return incorrectOrderIDOrPhotographerID
	}

	return nil
}

func (phr *PhotographerRepository) DelAllByOrderID(orderID int) error {
	store, err := phr.GetStore()
	if err != nil {
		return err
	}

	if _, err := store.db.Exec(
		`DELETE FROM "agreed_photographers" WHERE order_id = $1`, orderID); err != nil {
		return err
	}
	return nil
}

func (phr *PhotographerRepository) DelPhotographerByOrderID(photographerID int, orderID int) error {
	store, err := phr.GetStore()
	if err != nil {
		return err
	}

	if _, err := store.db.Exec(
		`DELETE FROM "agreed_photographers" WHERE photographer_id = $1`+
			` AND order_id = $2`, photographerID, orderID); err != nil {
		return err
	}
	return nil
}

func (phr *PhotographerRepository) GetStore() (*Store, error) {
	if phr.store == nil {
		return nil, errors.New("empty photographer store")
	}
	return phr.store, nil
}
