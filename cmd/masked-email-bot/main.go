package main

import (
	"context"
	"github.com/L11R/masked-email-bot/internal/domain"
	"github.com/L11R/masked-email-bot/internal/infra/fastmail"
	"github.com/L11R/masked-email-bot/internal/infra/sqlite"
	"github.com/L11R/masked-email-bot/internal/infra/telegram"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	TelegramConfig *telegram.Config
	DatabaseConfig *sqlite.Config
}

func main() {
	// Init logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Process config from environment variables
	var c Config
	if err := envconfig.Process(context.Background(), &c); err != nil {
		logger.Fatal("Cannot process config from env!", zap.Error(err))
	}

	db, err := sqlite.NewAdapter(logger, c.DatabaseConfig)
	if err != nil {
		logger.Fatal("Cannot init SQLite database!", zap.Error(err))
	}

	// Init fastmail client
	fmc := fastmail.NewAdapter(logger)

	// Init service
	service := domain.NewService(db, fmc)

	// Init telegram bot
	tgClient, err := telegram.NewAdapter(logger, c.TelegramConfig, service)

	// Setup graceful shutdown
	shutdown := make(chan error, 1)

	go func(shutdown chan<- error) {
		shutdown <- tgClient.Start()
	}(shutdown)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-sig:
		logger.Info("Got the signal!", zap.String("signal", s.String()))
		tgClient.Stop()
		db.Close()
	case err := <-shutdown:
		logger.Error("Error running the application!", zap.Error(err))
	}
}
