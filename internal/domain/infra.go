package domain

import (
	"context"
	"golang.org/x/oauth2"
)

type Database interface {
	CreateUser(telegramID int64, languageCode string) error
	UpdateToken(telegramID int64, fastmailToken string) error
	UpdateLanguageCode(telegramID int64, languageCode string) error
	GetUser(telegramID int64) (*User, error)

	CreateOAuth2State(state, codeVerifier string, telegramID int64) error
	GetOAuth2State(state string) (*OAuth2State, error)

	Close() error
}

type MaskingEmail interface {
	CreateMaskedEmail(token *oauth2.Token, forDomain string) (string, error)
	GetOAuth2Config() *oauth2.Config
}

type Delivery interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type Telegram interface {
	SendMessage(telegramID int64, languageCode, messageID string) error
}
