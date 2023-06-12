package telegram

import (
	"context"
	"github.com/L11R/masked-email-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
)

type delivery struct {
	logger  *zap.Logger
	config  *Config
	bundle  *i18n.Bundle
	bot     *tgbotapi.BotAPI
	service domain.Service
}

func NewDelivery(logger *zap.Logger, config *Config, bundle *i18n.Bundle, service domain.Service) (domain.Delivery, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	bot.Debug = config.Debug

	return &delivery{
		logger:  logger,
		config:  config,
		bundle:  bundle,
		bot:     bot,
		service: service,
	}, nil
}

func (d *delivery) ListenAndServe() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := d.bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		localizer := i18n.NewLocalizer(d.bundle, update.Message.From.LanguageCode)

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				if err := d.startCommand(localizer, update); err != nil {
					d.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			default:
				if err := d.anyOtherCommand(localizer, update); err != nil {
					d.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			}
		}

		if err := d.link(localizer, update); err != nil {
			d.logger.Error("Error while handling a link!", zap.Error(err))
		}
	}

	return nil
}

func (d *delivery) Shutdown(_ context.Context) error {
	d.bot.StopReceivingUpdates()
	return nil
}
