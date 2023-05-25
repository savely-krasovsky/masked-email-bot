package telegram

import (
	"github.com/L11R/masked-email-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type adapter struct {
	logger  *zap.Logger
	config  *Config
	bot     *tgbotapi.BotAPI
	service domain.Service
}

func NewAdapter(logger *zap.Logger, config *Config, service domain.Service) (domain.Bot, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	bot.Debug = config.Debug

	return &adapter{
		logger:  logger,
		config:  config,
		bot:     bot,
		service: service,
	}, nil
}

func (a *adapter) Start() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := a.bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				if err := a.startCommand(update); err != nil {
					a.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			case "token":
				if err := a.tokenCommand(update); err != nil {
					a.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			default:
				if err := a.anyOtherCommand(update); err != nil {
					a.logger.Error("Error while handling command!", zap.Error(err))
				}
				continue
			}
		}

		if err := a.link(update); err != nil {
			a.logger.Error("Error while handling a link!", zap.Error(err))
		}
	}

	return nil
}

func (a *adapter) Stop() error {
	a.bot.StopReceivingUpdates()
	return nil
}
