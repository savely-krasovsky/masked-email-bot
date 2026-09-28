package fastmail

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/L11R/masked-email-bot/internal/domain"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestCreateMaskedEmailResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantErr  error
	}{
		{
			name: "reserved prefix from production",
			// Fastmail returns no properties field for a reserved prefix.
			response: `{"methodResponses":[["MaskedEmail/set",{"notCreated":{"k1":{"type":"invalidProperties","description":"Name is reserved"}}},"0"]]}`,
			wantErr:  domain.ErrFastmailPrefixReserved,
		},
		{
			name:     "invalid prefix syntax",
			response: `{"methodResponses":[["MaskedEmail/set",{"notCreated":{"k1":{"type":"invalidProperties","description":"Invalid emailPrefix","properties":["emailPrefix"]}}},"0"]]}`,
			wantErr:  domain.ErrFastmailInternal,
		},
		{
			name:     "other error type with same description",
			response: `{"methodResponses":[["MaskedEmail/set",{"notCreated":{"k1":{"type":"forbidden","description":"Name is reserved"}}},"0"]]}`,
			wantErr:  domain.ErrFastmailInternal,
		},
		{
			name:     "rate limit",
			response: `{"methodResponses":[["MaskedEmail/set",{"notCreated":{"k1":{"type":"rateLimit"}}},"0"]]}`,
			wantErr:  domain.ErrFastmailInternal,
		},
		{
			name:     "different creation id",
			response: `{"methodResponses":[["MaskedEmail/set",{"notCreated":{"other":{"type":"invalidProperties","description":"Name is reserved"}}},"0"]]}`,
			wantErr:  domain.ErrFastmailInternal,
		},
		{
			name:     "method error",
			response: `{"methodResponses":[["error",{"type":"serverFail"},"0"]]}`,
			wantErr:  domain.ErrFastmailInternal,
		},
		{name: "no responses", response: `{"methodResponses":[]}`, wantErr: domain.ErrFastmailInternal},
		{name: "null response", response: `{"methodResponses":[null]}`, wantErr: domain.ErrFastmailInternal},
		{name: "null body", response: `{"methodResponses":[["MaskedEmail/set",null,"0"]]}`, wantErr: domain.ErrFastmailInternal},
		{name: "null created", response: `{"methodResponses":[["MaskedEmail/set",{"created":{"k1":null}},"0"]]}`, wantErr: domain.ErrFastmailInternal},
		{
			name:     "created",
			response: `{"methodResponses":[["MaskedEmail/set",{"created":{"k1":{"id":"masked-id","email":"example.abcde@fastmail.com"}}},"0"]]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tt.response)), Header: make(http.Header)}, nil
			})}
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
			a := &adapter{logger: zap.NewNop()}
			email, err := a.createMaskedEmail(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}), "account", "", "github")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if email == nil || email.ID != "masked-id" || email.Email != "example.abcde@fastmail.com" {
					t.Fatalf("unexpected created email: %+v", email)
				}
			} else if email != nil {
				t.Fatalf("unexpected email on failure: %+v", email)
			}
		})
	}
}

func TestReservedPrefixIsNotReplaced(t *testing.T) {
	for _, prefix := range []string{"github", "fastmail"} {
		for _, fromURL := range []bool{false, true} {
			name := prefix + "/prefix"
			if fromURL {
				name = prefix + "/url"
			}
			t.Run(name, func(t *testing.T) {
				var gotPrefix, gotDomain string
				client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					body := `{"primaryAccounts":{"https://www.fastmail.com/dev/maskedemail":"account"}}`
					if req.Method == http.MethodPost {
						var request Request[*MaskedEmailSetRequest]
						if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
							t.Fatal(err)
						}
						created := request.MethodCalls[0].Body.Create["k1"]
						gotPrefix, gotDomain = created.EmailPrefix, created.ForDomain
						body = `{"methodResponses":[["MaskedEmail/set",{"notCreated":{"k1":{"type":"invalidProperties","description":"Name is reserved"}}},"0"]]}`
					}
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
				})}
				ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
				a := NewAdapter(zap.NewNop(), &Config{})
				token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})
				var err error
				wantDomain := ""
				if fromURL {
					wantDomain = "https://" + prefix + ".com"
					u, parseErr := url.Parse(wantDomain + "/path?query=1#fragment")
					if parseErr != nil {
						t.Fatal(parseErr)
					}
					_, err = a.CreateMaskedEmailFromURL(ctx, token, u)
				} else {
					_, err = a.CreateMaskedEmailWithPrefix(ctx, token, prefix)
				}
				if !errors.Is(err, domain.ErrFastmailPrefixReserved) {
					t.Fatalf("error = %v, want reserved prefix", err)
				}
				if gotPrefix != prefix || gotDomain != wantDomain {
					t.Fatalf("sent prefix/domain = %q/%q, want %q/%q", gotPrefix, gotDomain, prefix, wantDomain)
				}
			})
		}
	}
}
