package service

import "errors"

var (
	ErrInvalidEventType     = errors.New("only spot_freed events are supported")
	ErrNotificationNotFound = errors.New("notification not found")
)
