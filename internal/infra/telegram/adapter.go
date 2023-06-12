package telegram

import (
	"github.com/L11R/masked-email-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
)

type adapter struct {
	logger *zap.Logger
	config *Config
	bundle *i18n.Bundle
	bot    *tgbotapi.BotAPI
}

func NewAdapter(logger *zap.Logger, config *Config, bundle *i18n.Bundle) (domain.Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	bot.Debug = config.Debug

	return &adapter{
		logger: logger,
		config: config,
		bundle: bundle,
		bot:    bot,
	}, nil
}

func (a *adapter) SendMessage(telegramID int64, languageCode, messageID string) error {
	localizer := i18n.NewLocalizer(a.bundle, languageCode)

	msg := tgbotapi.NewMessage(telegramID, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: messageID,
	}))
	msg.ParseMode = "MarkdownV2"

	if _, err := a.bot.Send(msg); err != nil {
		a.logger.Error("Error while sending a message!", zap.Error(err))
		return domain.ErrTelegramInternal
	}

	return nil
}
