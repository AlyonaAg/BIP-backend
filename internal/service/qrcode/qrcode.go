package qrcode

import (
	"strconv"
	"strings"

	"github.com/skip2/go-qrcode"

	"BIP_backend/internal/app/model"
)

type QRCoder struct {
	config *Config
}

func NewQRCoder() (*QRCoder, error) {
	configQRCoder, err := NewConfig()
	if err != nil {
		return nil, err
	}

	return &QRCoder{
		config: configQRCoder,
	}, nil
}

func (qc *QRCoder) CreateQRCode(location *model.Location, orderID int, secret string) ([]byte, error) {
	var png []byte

	info := qc.convertingInformationToString(location, orderID, secret)
	png, err := qrcode.Encode(info, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	return png, err
}

func (qc *QRCoder) DecodeQRCode(qrCode string) (*model.Location, int /*orderID*/, string /*secret*/, error) {
	qrCodeSplit := strings.Split(qrCode, "_")

	longitude, err := strconv.ParseFloat(qrCodeSplit[0], 64)
	if err != nil {
		return nil, 0, "", err
	}
	latitude, err := strconv.ParseFloat(qrCodeSplit[1], 64)
	if err != nil {
		return nil, 0, "", err
	}

	var location = &model.Location{
		Latitude:  latitude,
		Longitude: longitude,
	}

	orderID, err := strconv.Atoi(qrCodeSplit[2])
	if err != nil {
		return nil, 0, "", err
	}

	return location, orderID, qrCodeSplit[3], nil
}

func (qc *QRCoder) convertingInformationToString(location *model.Location, orderID int, secret string) string {
	return locationToString(location) + "_" + strconv.Itoa(orderID) + "_" + secret
}

func locationToString(location *model.Location) string {
	stringLongitude := strconv.FormatFloat(location.Longitude, 'f', -1, 64)
	stringLatitude := strconv.FormatFloat(location.Latitude, 'f', -1, 64)
	return stringLongitude + "_" + stringLatitude
}
