package domain

import "golang.org/x/oauth2"

type User struct {
	TelegramID    int64
	FastmailToken *oauth2.Token
	LanguageCode  string
}

type OAuth2State struct {
	State        string
	CodeVerifier string
	TelegramID   int64
}
