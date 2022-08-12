package store

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"BIP_backend/internal/app/model"
)

var (
	incorrectOrderID                 = errors.New("incorrect order id")
	incorrectOrderIDOrPhotographerID = errors.New("incorrect order id or photographer")
	incorrectQRCode                  = errors.New("incorrect CR-code")
	photosAreNotReadyYet             = errors.New("photos are not ready yet")
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

	o.OrderState = model.Created
	if err := store.db.QueryRow(
		`INSERT INTO "order" (client_id, order_cost, order_state, location, comment) `+
			`VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		o.ClientID, o.OrderCost, o.OrderState, locationToString(&o.Location), o.Comment,
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

	rows, err := store.db.Query(`SELECT id, client_id, photographer_id, order_cost, location, comment, ` +
		`order_state FROM "order" WHERE order_state = 'created' OR order_state = 'agreed_photographer'`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders, err := getOrders(rows)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (or *OrderRepository) GetClientOrders(clientID int) ([]model.Order, /*backlog*/
	[]model.Order, /*active*/
	[]model.Order, /*finished*/
	error) {
	store, err := or.GetStore()
	if err != nil {
		return nil, nil, nil, err
	}

	backlogRows, err := store.db.Query(`SELECT id, client_id, photographer_id, order_cost, location, comment, `+
		`order_state FROM "order" WHERE (order_state = 'created' OR order_state = 'agreed_photographer') AND client_id = $1`,
		clientID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer backlogRows.Close()
	backlog, err := getOrders(backlogRows)
	if err != nil {
		return nil, nil, nil, err
	}

	activeRows, err := store.db.Query(`SELECT id, client_id, photographer_id, order_cost, location, comment, `+
		`order_state FROM "order" WHERE (order_state = 'meeting' OR order_state = 'watermarks_sent' OR `+
		`order_state = 'agreed_client') AND client_id = $1`,
		clientID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer activeRows.Close()
	active, err := getOrders(activeRows)
	if err != nil {
		return nil, nil, nil, err
	}

	finishedRows, err := store.db.Query(`SELECT id, client_id, photographer_id, order_cost, location, comment, `+
		`order_state FROM "order" WHERE order_state = 'finish' AND client_id = $1`,
		clientID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer finishedRows.Close()
	finished, err := getOrders(finishedRows)
	if err != nil {
		return nil, nil, nil, err
	}

	return backlog, active, finished, nil
}

func (or *OrderRepository) GetPhotographerOrders(clientID int) ([]model.Order, /*backlog*/
	[]model.Order, /*active*/
	[]model.Order, /*finished*/
	error) {
	store, err := or.GetStore()
	if err != nil {
		return nil, nil, nil, err
	}

	backlogRows, err := store.db.Query(`SELECT id, client_id, photographer_id, order_cost, location, comment, `+
		`order_state FROM "order" WHERE order_state = 'agreed_photographer' AND $1 IN (SELECT DISTINCT photographer_id FROM `+
		`"agreed_photographers" WHERE "order".id = "agreed_photographers".order_id)`,
		clientID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer backlogRows.Close()
	backlog, err := getOrders(backlogRows)
	if err != nil {
		return nil, nil, nil, err
	}

	activeRows, err := store.db.Query(`SELECT id, client_id, photographer_id, order_cost, location, comment, `+
		`order_state FROM "order" WHERE (order_state = 'meeting' OR order_state = 'watermarks_sent' OR `+
		`order_state = 'agreed_client') AND photographer_id = $1`,
		clientID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer activeRows.Close()
	active, err := getOrders(activeRows)
	if err != nil {
		return nil, nil, nil, err
	}

	finishedRows, err := store.db.Query(`SELECT id, client_id, photographer_id, order_cost, location, comment, `+
		`order_state FROM "order" WHERE order_state = 'finish' AND photographer_id = $1`,
		clientID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer finishedRows.Close()
	finished, err := getOrders(finishedRows)
	if err != nil {
		return nil, nil, nil, err
	}

	return backlog, active, finished, nil
}

func (or *OrderRepository) GetOrderByID(orderID int) (*model.Order, error) {
	store, err := or.GetStore()
	if err != nil {
		return nil, err
	}

	var order = &model.Order{}
	var photographerID sql.NullInt64
	if err := store.db.QueryRow(
		`SELECT id, client_id, photographer_id, order_cost, `+
			` order_state, comment FROM "order" WHERE id = $1`,
		orderID,
	).Scan(
		&order.ID,
		&order.ClientID,
		&photographerID,
		&order.OrderCost,
		&order.OrderState,
		&order.Comment,
	); err != nil {
		return nil, err
	}

	if photographerID.Valid {
		order.PhotographerID = int(photographerID.Int64)
	}

	return order, nil
}

func (or *OrderRepository) GetClientIDByQRInfo(location *model.Location, orderID, photographerID int) (int, int, error) {
	store, err := or.GetStore()
	if err != nil {
		return 0, 0, err
	}

	var clientID, orderCost int
	var clientLocation string
	if err := store.db.QueryRow(
		`SELECT client_id, order_cost, client_current_location FROM "order" `+
			`WHERE id = $1 AND photographer_id = $2 AND order_state = 'agreed_client'`,
		orderID, photographerID,
	).Scan(
		&clientID, &orderCost, &clientLocation,
	); err != nil {
		return 0, 0, err
	}

	if clientLocation != locationToString(location) {
		return 0, 0, incorrectQRCode
	}

	return clientID, orderCost, nil
}

func (or *OrderRepository) GetURLOriginal(orderID int) (string, error) {
	store, err := or.GetStore()
	if err != nil {
		return "", err
	}

	var URLOriginal sql.NullString
	if err := store.db.QueryRow(
		`SELECT url_original FROM "order" WHERE id = $1`,
		orderID,
	).Scan(
		&URLOriginal,
	); err != nil {
		return "", err
	}

	if !URLOriginal.Valid {
		return "", photosAreNotReadyYet
	}
	return URLOriginal.String, nil
}

func (or *OrderRepository) GetURLWatermark(orderID int) (string, error) {
	store, err := or.GetStore()
	if err != nil {
		return "", err
	}

	var URLWatermark sql.NullString
	if err := store.db.QueryRow(
		`SELECT url_watermark FROM "order" WHERE id = $1`,
		orderID,
	).Scan(
		&URLWatermark,
	); err != nil {
		return "", err
	}

	if !URLWatermark.Valid {
		return "", photosAreNotReadyYet
	}

	return URLWatermark.String, nil
}

func (or *OrderRepository) GetPhotographerID(orderID int) (int, error) {
	store, err := or.GetStore()
	if err != nil {
		return 0, err
	}

	var photographerID int
	if err := store.db.QueryRow(
		`SELECT photographer_id FROM "order" WHERE id = $1`,
		orderID,
	).Scan(
		&photographerID,
	); err != nil {
		return 0, err
	}
	return photographerID, nil
}

func (or *OrderRepository) GetClientID(orderID int) (int, error) {
	store, err := or.GetStore()
	if err != nil {
		return 0, err
	}

	var clientID int
	if err := store.db.QueryRow(
		`SELECT client_id FROM "order" WHERE id = $1`,
		orderID,
	).Scan(
		&clientID,
	); err != nil {
		return 0, err
	}
	return clientID, nil
}

func (or *OrderRepository) GetCost(orderID int) (int, error) {
	store, err := or.GetStore()
	if err != nil {
		return 0, err
	}

	var cost int
	if err := store.db.QueryRow(
		`SELECT order_cost FROM "order" WHERE id = $1`,
		orderID,
	).Scan(&cost); err != nil {
		return 0, err
	}
	return cost, nil
}

func (or *OrderRepository) GetState(orderID int) (string, error) {
	store, err := or.GetStore()
	if err != nil {
		return "", err
	}

	var state string
	if err := store.db.QueryRow(
		`SELECT order_state FROM "order" WHERE id = $1`,
		orderID,
	).Scan(&state); err != nil {
		return "", err
	}
	return state, nil
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

func (or *OrderRepository) CheckMatchingClientAndOrderID(orderID, userID int) error {
	store, err := or.GetStore()
	if err != nil {
		return err
	}

	rows, err := store.db.Exec(
		`SELECT * FROM "order" WHERE id = $1 AND client_id = $2`, orderID, userID)
	if err != nil {
		return err
	}
	if count, _ := rows.RowsAffected(); count == 0 {
		return incorrectOrderID
	}

	return nil
}

func (or *OrderRepository) CheckMatchingPhotographAndOrderID(orderID, userID int) error {
	store, err := or.GetStore()
	if err != nil {
		return err
	}

	rows, err := store.db.Exec(
		`SELECT * FROM "order" WHERE id = $1 AND photographer_id = $2`, orderID, userID)
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

func (or *OrderRepository) UpdateCurrentLocation(location *model.Location, orderID int) error {
	store, err := or.GetStore()
	if err != nil {
		return err
	}

	rows, err := store.db.Exec(
		`UPDATE "order" SET client_current_location = $1 WHERE id = $2`,
		locationToString(location), orderID)
	if err != nil {
		return err
	}
	if count, _ := rows.RowsAffected(); count == 0 {
		return incorrectOrderID
	}

	return nil
}

func (or *OrderRepository) UpdateURL(URLOriginal, URLWatermark string, orderID int) error {
	store, err := or.GetStore()
	if err != nil {
		return err
	}

	rows, err := store.db.Exec(
		`UPDATE "order" SET url_original = $1, url_watermark = $2 WHERE id = $3`,
		URLOriginal, URLWatermark, orderID)
	if err != nil {
		return err
	}
	if count, _ := rows.RowsAffected(); count == 0 {
		return incorrectOrderID
	}

	return nil
}

func (or *OrderRepository) GetStore() (*Store, error) {
	if or.store == nil {
		return nil, errors.New("empty order store")
	}
	return or.store, nil
}

func locationToString(location *model.Location) string {
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

func getOrders(rows *sql.Rows) ([]model.Order, error) {
	var orders []model.Order
	var photographerID sql.NullInt64
	var locationString string

	for rows.Next() {
		var order = model.Order{}
		if err := rows.Scan(
			&order.ID,
			&order.ClientID,
			&photographerID,
			&order.OrderCost,
			&locationString,
			&order.Comment,
			&order.OrderState,
		); err != nil {
			return nil, err
		}

		location, err := stringToLocation(locationString)
		if err != nil {
			continue
		}

		if photographerID.Valid {
			order.PhotographerID = int(photographerID.Int64)
		}

		order.Location.Latitude = location.Latitude
		order.Location.Longitude = location.Longitude

		orders = append(orders, order)
	}
	return orders, nil
}
