package service

import "errors"

var (
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrUserNotFound              = errors.New("user not found")
	ErrZoneNotFound              = errors.New("zone not found")
	ErrDependencyUnavailable     = errors.New("dependency service unavailable")
)
