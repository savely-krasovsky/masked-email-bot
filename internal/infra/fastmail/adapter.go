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
	"regexp"
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

func (a *adapter) openSession(ctx context.Context, tokenSrc oauth2.TokenSource) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.fastmail.com/jmap/session", nil)
	if err != nil {
		a.logger.Error("Error while creating a new HTTP request!", zap.Error(err))
		return "", domain.ErrFastmailInternal
	}

	resp, err := oauth2.NewClient(ctx, tokenSrc).Do(req)
	if err != nil {
		a.logger.Error("Error while doing an HTTP request!", zap.Error(err))
		return "", domain.ErrFastmailInternal
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
		return "", domain.ErrFastmailInternal
	}

	for k, v := range jsonResp.PrimaryAccounts {
		if k == "https://www.fastmail.com/dev/maskedemail" {
			return v, nil
		}
	}

	return "", domain.ErrFastmailPrimaryAccountNotFound
}

func (a *adapter) createMaskedEmail(ctx context.Context, tokenSrc oauth2.TokenSource, accountID, forDomain, emailPrefix string) (*MaskedEmail, error) {
	request := &Request[*MaskedEmailSetRequest]{
		Using: []string{"https://www.fastmail.com/dev/maskedemail"},
		MethodCalls: []*Invocation[*MaskedEmailSetRequest]{
			{
				Name: "MaskedEmail/set",
				Body: &MaskedEmailSetRequest{
					AccountID: accountID,
					Create: map[string]*MaskedEmail{
						"k1": {
							ForDomain:   forDomain,
							EmailPrefix: emailPrefix,
						},
					},
				},
				ID: "0",
			},
		},
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(request); err != nil {
		a.logger.Error("Error while trying to encode JSON request!", zap.Error(err))
		return nil, domain.ErrFastmailInternal
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.fastmail.com/jmap/api/", buf)
	if err != nil {
		a.logger.Error("Error while creating a new HTTP request!", zap.Error(err))
		return nil, domain.ErrFastmailInternal
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauth2.NewClient(ctx, tokenSrc).Do(req)
	if err != nil {
		a.logger.Error("Error while doing an HTTP request!", zap.Error(err))
		return nil, domain.ErrFastmailInternal
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.logger.Error("Wrong status code!")
		return nil, domain.ErrFastmailInternal
	}

	var jsonResp struct {
		MethodResponses []*Invocation[*MaskedEmailSetResponse] `json:"methodResponses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		a.logger.Error("Error while trying to decode JSON response!", zap.Error(err))
		return nil, domain.ErrFastmailInternal
	}

	created, ok := jsonResp.MethodResponses[0].Body.Created["k1"]
	if !ok {
		return nil, domain.ErrFastmailInternal
	}

	return created, nil
}

func (a *adapter) CreateMaskedEmailFromURL(ctx context.Context, tokenSrc oauth2.TokenSource, u *url.URL) (*domain.MaskedEmail, error) {
	u.Opaque = ""
	u.User = nil
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""

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

	// remove all special characters except underscore
	emailPrefix = regexp.MustCompile(`[^a-zA-Z0-9_]+`).ReplaceAllString(emailPrefix, "")

	accountId, err := a.openSession(ctx, tokenSrc)
	if err != nil {
		return nil, err
	}

	maskedEmail, err := a.createMaskedEmail(ctx, tokenSrc, accountId, u.String(), emailPrefix)
	if err != nil {
		return nil, err
	}

	return &domain.MaskedEmail{
		ID:    maskedEmail.ID,
		Email: maskedEmail.Email,
	}, nil
}

func (a *adapter) CreateMaskedEmailWithPrefix(ctx context.Context, tokenSrc oauth2.TokenSource, prefix string) (*domain.MaskedEmail, error) {
	accountId, err := a.openSession(ctx, tokenSrc)
	if err != nil {
		return nil, err
	}

	maskedEmail, err := a.createMaskedEmail(ctx, tokenSrc, accountId, "", prefix)
	if err != nil {
		return nil, err
	}

	return &domain.MaskedEmail{
		ID:    maskedEmail.ID,
		Email: maskedEmail.Email,
	}, nil
}

func (a *adapter) enableMaskedEmail(ctx context.Context, tokenSrc oauth2.TokenSource, accountID, id string) error {
	request := &Request[*MaskedEmailSetRequest]{
		Using: []string{"https://www.fastmail.com/dev/maskedemail"},
		MethodCalls: []*Invocation[*MaskedEmailSetRequest]{
			{
				Name: "MaskedEmail/set",
				Body: &MaskedEmailSetRequest{
					AccountID: accountID,
					Update: map[string]*MaskedEmail{
						id: {
							State: MaskedEmailStateEnabled,
						},
					},
				},
				ID: "0",
			},
		},
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(request); err != nil {
		a.logger.Error("Error while trying to encode JSON request!", zap.Error(err))
		return domain.ErrFastmailInternal
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.fastmail.com/jmap/api/", buf)
	if err != nil {
		a.logger.Error("Error while creating a new HTTP request!", zap.Error(err))
		return domain.ErrFastmailInternal
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauth2.NewClient(ctx, tokenSrc).Do(req)
	if err != nil {
		a.logger.Error("Error while doing an HTTP request!", zap.Error(err))
		return domain.ErrFastmailInternal
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.logger.Error("Wrong status code!")
		return domain.ErrFastmailInternal
	}

	var jsonResp struct {
		MethodResponses []*Invocation[*MaskedEmailSetResponse] `json:"methodResponses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		a.logger.Error("Error while trying to decode JSON response!", zap.Error(err))
		return domain.ErrFastmailInternal
	}

	return nil
}

func (a *adapter) EnableMaskedEmail(ctx context.Context, tokenSrc oauth2.TokenSource, id string) error {
	accountId, err := a.openSession(ctx, tokenSrc)
	if err != nil {
		return err
	}

	if err := a.enableMaskedEmail(ctx, tokenSrc, accountId, id); err != nil {
		return err
	}

	return nil
}
