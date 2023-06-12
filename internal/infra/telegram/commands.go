package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
)

func (d *delivery) startCommand(localizer *i18n.Localizer, update tgbotapi.Update) error {
	authCodeURL, err := d.service.StartCommand(update.Message.From.ID, update.Message.From.LanguageCode)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "TelegramError",
		}))
		if _, err := d.bot.Send(msg); err != nil {
			d.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "TelegramStartCommand",
	}))
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
		{
			Text: localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "TelegramStartCommandAuthButton"}),
			URL:  &authCodeURL,
		},
	})

	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (d *delivery) anyOtherCommand(localizer *i18n.Localizer, update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.From.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "TelegramUnknownCommand",
	}))
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (d *delivery) link(localizer *i18n.Localizer, update tgbotapi.Update) error {
	email, err := d.service.Link(update.Message.From.ID, update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "TelegramError",
		}))
		if _, err := d.bot.Send(msg); err != nil {
			d.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "TelegramEmail",
		TemplateData: map[string]interface{}{
			"Email": email,
		},
	}))
	msg.ParseMode = "MarkdownV2"
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}
