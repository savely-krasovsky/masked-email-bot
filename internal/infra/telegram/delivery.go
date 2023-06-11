package telegram

import (
	"context"
	"github.com/L11R/masked-email-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type delivery struct {
	logger  *zap.Logger
	config  *Config
	bot     *tgbotapi.BotAPI
	service domain.Service
}

func NewDelivery(logger *zap.Logger, config *Config, service domain.Service) (domain.Delivery, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	bot.Debug = config.Debug

	return &delivery{
		logger:  logger,
		config:  config,
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

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				if err := d.startCommand(update); err != nil {
					d.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			case "token":
				if err := d.tokenCommand(update); err != nil {
					d.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			case "auth":
				if err := d.authCommand(update); err != nil {
					d.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			default:
				if err := d.anyOtherCommand(update); err != nil {
					d.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			}
		}

		if err := d.link(update); err != nil {
			d.logger.Error("Error while handling a link!", zap.Error(err))
		}
	}

	return nil
}

func (d *delivery) Shutdown(_ context.Context) error {
	d.bot.StopReceivingUpdates()
	return nil
}
