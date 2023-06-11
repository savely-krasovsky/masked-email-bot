package fastmail

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/L11R/masked-email-bot/internal/domain"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
)

type adapter struct {
	logger *zap.Logger
	config *Config
}

func NewAdapter(logger *zap.Logger, config *Config) domain.MaskingEmail {
	return &adapter{
		logger: logger,
		config: config,
	}
}

func (a *adapter) openSession(token *oauth2.Token) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.fastmail.com/jmap/session", nil)
	if err != nil {
		a.logger.Error("Error while creating a new HTTP request!", zap.Error(err))
		return "", err
	}

	resp, err := a.GetOAuth2Config().Client(context.Background(), token).Do(req)
	if err != nil {
		a.logger.Error("Error while doing an HTTP request!", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.logger.Error("Wrong status code!")
		return "", domain.ErrFastmailInternal
	}

	var jsonResp struct {
		PrimaryAccounts map[string]string `json:"primaryAccounts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		a.logger.Error("Error while trying to decode JSON response!", zap.Error(err))
		return "", err
	}

	for k, v := range jsonResp.PrimaryAccounts {
		if k == "https://www.fastmail.com/dev/maskedemail" {
			return v, nil
		}
	}

	return "", domain.ErrFastmailPrimaryAccountNotFound
}

func (a *adapter) createMaskedEmail(token *oauth2.Token, accountID, forDomain, emailPrefix string) (string, error) {
	request := struct {
		Using       []string `json:"using"`
		MethodCalls []any    `json:"methodCalls"`
	}{
		Using: []string{"https://www.fastmail.com/dev/maskedemail"},
		MethodCalls: []any{
			[]any{
				"MaskedEmail/set",
				map[string]any{
					"accountId": accountID,
					"create": map[string]any{
						"k1": map[string]string{
							"forDomain":   forDomain,
							"emailPrefix": emailPrefix,
						},
					},
				},
				"0",
			},
		},
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(request); err != nil {
		a.logger.Error("Error while trying to encode JSON request!", zap.Error(err))
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.fastmail.com/jmap/api/", buf)
	if err != nil {
		a.logger.Error("Error while creating a new HTTP request!", zap.Error(err))
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.GetOAuth2Config().Client(context.Background(), token).Do(req)
	if err != nil {
		a.logger.Error("Error while doing an HTTP request!", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.logger.Error("Wrong status code!")
		return "", domain.ErrFastmailInternal
	}

	var jsonResp struct {
		MethodResponses [][]json.RawMessage `json:"methodResponses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		a.logger.Error("Error while trying to decode JSON response!", zap.Error(err))
		return "", err
	}

	var (
		methodName string
		body       struct {
			Created struct {
				K1 struct {
					Email string `json:"email"`
				} `json:"k1"`
			} `json:"created"`
		}
		status string
	)
	if len(jsonResp.MethodResponses) > 0 && len(jsonResp.MethodResponses[0]) > 0 {
		for index, raw := range jsonResp.MethodResponses[0] {
			switch index {
			case 0:
				if err := json.Unmarshal(raw, &methodName); err != nil {
					a.logger.Error("Error while trying to parse JSON response!", zap.Error(err))
					return "", err
				}
			case 1:
				if err := json.Unmarshal(raw, &body); err != nil {
					a.logger.Error("Error while trying to parse JSON response!", zap.Error(err))
					return "", err
				}
			case 2:
				if err := json.Unmarshal(raw, &status); err != nil {
					a.logger.Error("Error while trying to parse JSON response!", zap.Error(err))
					return "", err
				}
			}
		}
	}

	return body.Created.K1.Email, nil
}

func (a *adapter) CreateMaskedEmail(token *oauth2.Token, forDomain string) (string, error) {
	u, err := url.Parse(forDomain)
	if err != nil {
		return "", err
	}
	u.Path = ""
	u.RawQuery = ""

	emailPrefix := ""

	if _, err := netip.ParseAddr(u.Hostname()); err == nil {
		emailPrefix = "ipaddr"
	}

	parts := strings.Split(u.Hostname(), ".")
	switch len(parts) {
	case 1:
		emailPrefix = parts[0]
	case 2:
		emailPrefix = parts[0]
	default:
		emailPrefix = parts[len(parts)-2]
	}

	switch emailPrefix {
	case "fastmail":
		emailPrefix = "mail"
	case "github":
		emailPrefix = "dev"
	}

	accountId, err := a.openSession(token)
	if err != nil {
		return "", err
	}

	return a.createMaskedEmail(token, accountId, u.String(), emailPrefix)
}
