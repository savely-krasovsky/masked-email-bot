package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/L11R/masked-email-bot/internal/domain"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func testConfig(t *testing.T) *Config {
	t.Helper()
	var config Config
	if err := envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target: &config, Lookuper: envconfig.MapLookuper(map[string]string{}),
	}); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(config.DBFile)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = filepath.Join(t.TempDir(), "masked_email_bot.db")
	config.DBFile = u.String()
	migrations, err := filepath.Abs("../../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	config.MigrationsSourceURL = (&url.URL{Scheme: "file", Path: migrations}).String()
	return &config
}

func openTestAdapter(t *testing.T, config *Config) domain.Database {
	t.Helper()
	db, err := NewAdapter(zap.NewNop(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestAdapterMigrationsAndPersistence(t *testing.T) {
	config := testConfig(t)
	db := openTestAdapter(t, config)
	const telegramID = int64(123)

	if err := db.CreateUser(telegramID, "en"); err != nil {
		t.Fatal(err)
	}
	user, err := db.GetUser(telegramID)
	if err != nil {
		t.Fatal(err)
	}
	if user.FastmailToken != nil || user.LanguageCode != "en" {
		t.Fatalf("unexpected new user: %+v", user)
	}
	if err := db.CreateUser(telegramID, "ru"); !errors.Is(err, domain.ErrSqliteUserAlreadyExists) {
		t.Fatalf("duplicate user error = %v", err)
	}
	if err := db.UpdateLanguageCode(telegramID, "ru"); err != nil {
		t.Fatal(err)
	}

	token := oauth2.Token{
		AccessToken: "test-access", RefreshToken: "test-refresh", TokenType: "Bearer",
		Expiry: time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
	data, err := json.Marshal(token)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateToken(telegramID, string(data)); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateOAuth2State("test-state", "test-verifier", telegramID); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	// Reopening must accept the existing migration version and retain all data.
	db = openTestAdapter(t, config)
	user, err = db.GetUser(telegramID)
	if err != nil {
		t.Fatal(err)
	}
	if user.TelegramID != telegramID || user.LanguageCode != "ru" {
		t.Fatalf("unexpected persisted user: %+v", user)
	}
	if user.FastmailToken == nil || user.FastmailToken.AccessToken != token.AccessToken ||
		user.FastmailToken.RefreshToken != token.RefreshToken || user.FastmailToken.TokenType != token.TokenType ||
		!user.FastmailToken.Expiry.Equal(token.Expiry) {
		t.Fatalf("unexpected persisted OAuth token: %+v", user.FastmailToken)
	}
	state, err := db.GetOAuth2State("test-state")
	if err != nil {
		t.Fatal(err)
	}
	if state.State != "test-state" || state.CodeVerifier != "test-verifier" || state.TelegramID != telegramID {
		t.Fatalf("unexpected persisted OAuth state: %+v", state)
	}

	var version int
	var dirty bool
	if err := db.(*adapter).db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatal(err)
	}
	if version != 20230612161105 || dirty {
		t.Fatalf("unexpected migration state: version=%d dirty=%t", version, dirty)
	}
}

func TestAdapterMissingRecords(t *testing.T) {
	db := openTestAdapter(t, testConfig(t))
	if _, err := db.GetUser(123); !errors.Is(err, domain.ErrNoUser) {
		t.Fatalf("missing user error = %v, want %v", err, domain.ErrNoUser)
	}
	if _, err := db.GetOAuth2State("missing"); !errors.Is(err, domain.ErrNoState) {
		t.Fatalf("missing OAuth state error = %v, want %v", err, domain.ErrNoState)
	}
}

func TestDefaultBusyTimeoutOnEachConnection(t *testing.T) {
	db := openTestAdapter(t, testConfig(t)).(*adapter).db
	// Keep both connections open so the pool has to initialize a second one.
	for range 2 {
		conn, err := db.Conn(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		var timeout int
		if err := conn.QueryRowContext(context.Background(), "PRAGMA busy_timeout").Scan(&timeout); err != nil {
			t.Fatal(err)
		}
		if timeout != 5000 {
			t.Fatalf("busy_timeout = %d, want 5000 ms", timeout)
		}
	}
}
