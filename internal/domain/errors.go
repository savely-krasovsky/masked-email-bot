package domain

import "errors"

var (
	ErrNoUser                         = errors.New("common: no user")
	ErrNoToken                        = errors.New("common: no token")
	ErrFastmailInternal               = errors.New("fastmail: internal error")
	ErrFastmailPrimaryAccountNotFound = errors.New("fastmail: primary account not found")
	ErrTelegramInternal               = errors.New("telegram: internal error")
	ErrSqliteInternal                 = errors.New("sqlite: internal error")
)
