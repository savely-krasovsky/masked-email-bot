package main

import (
	"context"
	"embed"
	"github.com/BurntSushi/toml"
	"github.com/L11R/masked-email-bot/internal/domain"
	"github.com/L11R/masked-email-bot/internal/infra/fastmail"
	"github.com/L11R/masked-email-bot/internal/infra/httpserver"
	"github.com/L11R/masked-email-bot/internal/infra/sqlite"
	"github.com/L11R/masked-email-bot/internal/infra/telegram"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"io/fs"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	TelegramConfig *telegram.Config
	HTTPConfig     *httpserver.Config
	FastmailConfig *fastmail.Config
	DatabaseConfig *sqlite.Config
}

//go:embed locales/*.toml
var localeFS embed.FS

func main() {
	// Init logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Process config from environment variables
	var c Config
	if err := envconfig.Process(context.Background(), &c); err != nil {
		logger.Fatal("Cannot process config from env!", zap.Error(err))
	}

	// Init SQLite adapter
	db, err := sqlite.NewAdapter(logger, c.DatabaseConfig)
	if err != nil {
		logger.Fatal("Cannot init SQLite database!", zap.Error(err))
	}

	// Init Fastmail adapter
	fmc := fastmail.NewAdapter(logger, c.FastmailConfig)

	// Internalization (i18n)
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	fs.WalkDir(localeFS, "locales", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if _, err := bundle.LoadMessageFileFS(localeFS, path); err != nil {
			return err
		}

		return nil
	})

	// Init Telegram adapter
	t, err := telegram.NewAdapter(logger, c.TelegramConfig, bundle)
	if err != nil {
		logger.Fatal("Cannot init Telegram adapter!", zap.Error(err))
	}

	// Init service
	service := domain.NewService(logger, db, fmc, t)

	// Setup graceful shutdown
	shutdown := make(chan error, 1)

	// Init Telegram delivery
	telegramDelivery, err := telegram.NewDelivery(logger, c.TelegramConfig, bundle, service)
	if err != nil {
		logger.Fatal("Cannot init Telegram delivery!", zap.Error(err))
	}

	go func(shutdown chan<- error) {
		shutdown <- telegramDelivery.ListenAndServe()
	}(shutdown)

	// Init HTTP delivery
	httpDelivery, err := httpserver.NewDelivery(logger, c.HTTPConfig, service)
	if err != nil {
		logger.Fatal("Cannot init HTTP delivery!", zap.Error(err))
	}

	go func(shutdown chan<- error) {
		shutdown <- httpDelivery.ListenAndServe()
	}(shutdown)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-sig:
		logger.Info("Got the signal!", zap.String("signal", s.String()))
		telegramDelivery.Shutdown(nil)
		db.Close()
	case err := <-shutdown:
		logger.Error("Error running the application!", zap.Error(err))
	}
}
