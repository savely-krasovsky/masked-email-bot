package fastmail

import "golang.org/x/oauth2"

func (a *adapter) GetOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID: a.config.ClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:   a.config.AuthURL,
			TokenURL:  a.config.TokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: a.config.RedirectURL,
		Scopes:      a.config.Scopes,
	}
}
