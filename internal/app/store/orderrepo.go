package store

import (
	"errors"
	"strconv"
	"strings"

	"BIP_backend/internal/app/model"
)

var (
	incorrectOrderID                 = errors.New("incorrect order id")
	incorrectOrderIDOrPhotographerID = errors.New("incorrect order id or photographer")
)

type OrderRepository struct {
	store *Store
}

func (or *OrderRepository) Create(o *model.Order) error {
	if err := o.Validate(); err != nil {
		return err
	}

	store, err := or.GetStore()
	if err != nil {
		return err
	}

	if err := store.db.QueryRow(
		`INSERT INTO "order" (client_id, order_cost, order_state, location, comment) `+
			`VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		o.ClientID, o.OrderCost, o.OrderState, locationToString(o.Location), o.Comment,
	).Scan(
		&o.ID,
	); err != nil {
		return err
	}
	return nil
}

func (or *OrderRepository) GetListCreatedOrder() ([]model.Order, error) {
	store, err := or.GetStore()
	if err != nil {
		return nil, err
	}

	rows, err := store.db.Query(`SELECT id, client_id, order_cost, location, comment ` +
		`FROM "order" WHERE order_state = 'created' OR order_state = 'agreed_photographer'`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	var locationString string

	for rows.Next() {
		var order = model.Order{}
		if err := rows.Scan(
			&order.ID,
			&order.ClientID,
			&order.OrderCost,
			&locationString,
			&order.Comment,
		); err != nil {
			return nil, err
		}

		location, err := stringToLocation(locationString)
		if err != nil {
			return nil, err
		}
		order.Location.Latitude = location.Latitude
		order.Location.Longitude = location.Longitude

		orders = append(orders, order)
	}
	return orders, err
}

func (or *OrderRepository) CheckOrderAvailability(orderID int) error {
	store, err := or.GetStore()
	if err != nil {
		return err
	}

	rows, err := store.db.Exec(
		`SELECT * FROM "order" WHERE id = $1`, orderID)
	if err != nil {
		return err
	}
	if count, _ := rows.RowsAffected(); count == 0 {
		return incorrectOrderID
	}

	return nil
}

func (or *OrderRepository) UpdateOrderState(newState string, orderID int) error {
	store, err := or.GetStore()
	if err != nil {
		return err
	}

	rows, err := store.db.Exec(
		`UPDATE "order" SET order_state = $1 WHERE id = $2`, newState, orderID)
	if err != nil {
		return err
	}
	if count, _ := rows.RowsAffected(); count == 0 {
		return incorrectOrderID
	}

	return nil
}

func (or *OrderRepository) UpdateOrderPhotographer(photographerID, orderID int) error {
	store, err := or.GetStore()
	if err != nil {
		return err
	}

	rows, err := store.db.Exec(
		`UPDATE "order" SET photographer_id = $1 WHERE id = $2`, photographerID, orderID)
	if err != nil {
		return err
	}
	if count, _ := rows.RowsAffected(); count == 0 {
		return incorrectOrderIDOrPhotographerID
	}

	return nil
}

func (or *OrderRepository) GetStore() (*Store, error) {
	if or.store == nil {
		return nil, errors.New("empty order store")
	}
	return or.store, nil
}

func locationToString(location model.Location) string {
	stringLongitude := strconv.FormatFloat(location.Longitude, 'f', -1, 64)
	stringLatitude := strconv.FormatFloat(location.Latitude, 'f', -1, 64)
	return "(" + stringLongitude + "," + stringLatitude + ")"
}

func stringToLocation(locationString string) (*model.Location, error) {
	locationSlice := strings.Split(strings.Trim(locationString, "()"), ",")

	latitude, err := strconv.ParseFloat(locationSlice[0], 64)
	if err != nil {
		return nil, err
	}
	longitude, err := strconv.ParseFloat(locationSlice[1], 64)
	if err != nil {
		return nil, err
	}

	return &model.Location{
		Latitude:  latitude,
		Longitude: longitude,
	}, nil
}
