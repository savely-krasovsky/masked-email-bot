package telegram

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/L11R/masked-email-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

type failingEmailService struct {
	domain.Service
	err error
}

func (s failingEmailService) GenerateMaskedEmail(int64, string) (*domain.MaskedEmail, error) {
	return nil, s.err
}

func (s failingEmailService) Prefix(int64, string) (*domain.MaskedEmail, error) {
	return nil, s.err
}

type telegramHTTPClientFunc func(*http.Request) (*http.Response, error)

func (f telegramHTTPClientFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMaskedEmailErrorReplies(t *testing.T) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, lang := range []string{"en", "ru"} {
		if _, err := bundle.LoadMessageFile(filepath.Join("..", "..", "..", "cmd", "masked-email-bot", "locales", lang+".toml")); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name string
		lang string
		err  error
		want string
	}{
		{"reserved ru", "ru", domain.ErrFastmailPrefixReserved, "Этот префикс запрещён Fastmail. Выберите другой."},
		{"reserved en", "en", domain.ErrFastmailPrefixReserved, "This prefix is reserved by Fastmail. Please choose another one."},
		{"wrapped error", "ru", fmt.Errorf("create: %w", domain.ErrFastmailPrefixReserved), "Этот префикс запрещён Fastmail. Выберите другой."},
		{"fallback language", "de", domain.ErrFastmailPrefixReserved, "This prefix is reserved by Fastmail. Please choose another one."},
		{"other error ru", "ru", domain.ErrFastmailInternal, "Произошло нечто ужасное! Попробуйте снова позже..."},
		{"other error en", "en", domain.ErrFastmailInternal, "Something bad happened! Please, try again later..."},
	}
	for _, tt := range tests {
		for _, inline := range []bool{false, true} {
			name := tt.name + "/message"
			if inline {
				name = tt.name + "/inline"
			}
			t.Run(name, func(t *testing.T) {
				var got url.Values
				var method string
				var calls int
				client := telegramHTTPClientFunc(func(req *http.Request) (*http.Response, error) {
					body := `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test"}}`
					if !strings.HasSuffix(req.URL.Path, "/getMe") {
						calls++
						if err := req.ParseForm(); err != nil {
							t.Fatal(err)
						}
						got, method = req.PostForm, filepath.Base(req.URL.Path)
						body = `{"ok":true,"result":{}}`
						if inline {
							body = `{"ok":true,"result":true}`
						}
					}
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
				})
				bot, err := tgbotapi.NewBotAPIWithClient("test-token", "https://telegram.example/bot%s/%s", client)
				if err != nil {
					t.Fatal(err)
				}
				d := &delivery{logger: zap.NewNop(), bot: bot, service: failingEmailService{err: tt.err}}
				localizer := i18n.NewLocalizer(bundle, tt.lang)
				user := &tgbotapi.User{ID: 123}
				if inline {
					err = d.generateMaskedEmailWithInlineButton(localizer, tgbotapi.Update{CallbackQuery: &tgbotapi.CallbackQuery{ID: "callback-id", From: user, Data: "prefix:github"}})
					if method != "answerCallbackQuery" || got.Get("show_alert") != "true" || got.Get("callback_query_id") != "callback-id" {
						t.Fatalf("unexpected inline reply: %s %v", method, got)
					}
				} else {
					err = d.generateMaskedEmail(localizer, tgbotapi.Update{Message: &tgbotapi.Message{From: user, Text: "https://github.com"}})
					if method != "sendMessage" || got.Get("chat_id") != "123" {
						t.Fatalf("unexpected message reply: %s %v", method, got)
					}
				}
				if !errors.Is(err, tt.err) {
					t.Fatalf("error = %v, want %v", err, tt.err)
				}
				if calls != 1 || got.Get("text") != tt.want {
					t.Fatalf("got %d replies, text %q; want one reply with %q", calls, got.Get("text"), tt.want)
				}
			})
		}
	}
}
