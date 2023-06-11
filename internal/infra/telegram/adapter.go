package telegram

import (
	"github.com/L11R/masked-email-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type adapter struct {
	logger *zap.Logger
	config *Config
	bot    *tgbotapi.BotAPI
}

func NewAdapter(logger *zap.Logger, config *Config) (domain.Telegram, error) {
	a := &adapter{
		logger: logger,
		config: config,
	}

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	bot.Debug = config.Debug

	return a, nil
}

func (a *adapter) SendMessage(telegramID int64, text string) error {
	msg := tgbotapi.NewMessage(telegramID, text)
	if _, err := a.bot.Send(msg); err != nil {
		a.logger.Error("Error while sending a message!", zap.Error(err))
		return domain.ErrTelegramInternal
	}

	return nil
}
