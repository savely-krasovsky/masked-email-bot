package fastmail

type Config struct {
	ClientID    string   `env:"FASTMAIL_OAUTH2_CLIENT_ID,required"`
	RedirectURL string   `env:"FASTMAIL_OAUTH2_REDIRECT_URL,required"`
	AuthURL     string   `env:"FASTMAIL_OAUTH2_AUTH_URL,default=https://api.fastmail.com/oauth/authorize"`
	TokenURL    string   `env:"FASTMAIL_OAUTH2_TOKEN_URL,default=https://api.fastmail.com/oauth/refresh"`
	Scopes      []string `env:"FASTMAIL_OAUTH2_SCOPES,default=urn:ietf:params:jmap:core,https://www.fastmail.com/dev/maskedemail"`
}
