package domain

type Database interface {
	CreateUser(telegramID int) error
	UpdateToken(telegramID int, fastmailToken string) error
	GetUser(telegramID int) (*User, error)

	Close() error
}

type MaskingEmail interface {
	CreateMaskedEmail(token, forDomain string) (string, error)
}

type Bot interface {
	Start() error
	Stop() error
}
