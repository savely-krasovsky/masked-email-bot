package domain

import "errors"

var (
	ErrNoUser                         = errors.New("common: no user")
	ErrNoToken                        = errors.New("common: no token")
	ErrNoState                        = errors.New("common: no state")
	ErrRandom                         = errors.New("common: cannot generate random bytes")
	ErrJSONEncoding                   = errors.New("common: cannot encode json")
	ErrFastmailInternal               = errors.New("fastmail: internal error")
	ErrFastmailPrimaryAccountNotFound = errors.New("fastmail: primary account not found")
	ErrTelegramInternal               = errors.New("telegram: internal error")
	ErrHTTPInternal                   = errors.New("http: internal error")
	ErrSqliteInternal                 = errors.New("sqlite: internal error")
	ErrSqliteUserAlreadyExists        = errors.New("sqlite: user already exists")
)
