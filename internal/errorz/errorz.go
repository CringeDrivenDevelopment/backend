package errorz

import "errors"

var (
	ErrNotEnoughPerms  = errors.New("not enough permissions")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidInitData = errors.New("invalid init data")
)
