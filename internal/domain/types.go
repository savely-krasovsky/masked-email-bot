package domain

import "golang.org/x/oauth2"

type User struct {
	TelegramID    int
	FastmailToken *oauth2.Token
}

type OAuth2State struct {
	State        string
	CodeVerifier string
	TelegramID   int64
}
