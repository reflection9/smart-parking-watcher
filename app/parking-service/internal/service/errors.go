package service

import "errors"

var (
	ErrZoneAlreadyExists = errors.New("zone already exists")
	ErrZoneNotFound      = errors.New("zone not found")
	ErrSpotAlreadyExists = errors.New("spot already exists in zone")
	ErrSpotNotFound      = errors.New("spot not found")
	ErrInvalidSpotStatus = errors.New("invalid spot status")
)
