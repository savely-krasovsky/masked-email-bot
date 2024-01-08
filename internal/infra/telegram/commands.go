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
	maskedEmail, err := d.service.Link(update.Message.From.ID, update.Message.Text)
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
			"Email": maskedEmail.Email,
		},
	}))
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
		{
			Text:         localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "TelegramEmailDoNotDeleteButton"}),
			CallbackData: &maskedEmail.ID,
		},
	})
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (d *delivery) enableMaskedEmail(localizer *i18n.Localizer, update tgbotapi.Update) error {
	if err := d.service.EnableMaskedEmail(update.CallbackQuery.From.ID, update.CallbackData()); err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "TelegramError",
		}))
		callback.ShowAlert = true
		if _, err := d.bot.Request(callback); err != nil {
			d.logger.Error("Error while answering to the callback query!", zap.Error(err))
		}
		return err
	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "TelegramEmailActivated",
	}))
	if _, err := d.bot.Request(callback); err != nil {
		d.logger.Error("Error while answering to the callback query!", zap.Error(err))
	}

	// Remove inline reply markup
	/*if _, err := d.bot.Send(&tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    update.CallbackQuery.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.MessageID,
		},
	}); err != nil {
		d.logger.Error("Error while removing inline keyboard!", zap.Error(err))
	}*/

	runes := []rune(update.CallbackQuery.Message.Text)
	start := update.CallbackQuery.Message.Entities[0].Offset
	end := update.CallbackQuery.Message.Entities[0].Offset + update.CallbackQuery.Message.Entities[0].Length
	email := string(runes[start:end])

	msg := tgbotapi.NewEditMessageText(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "TelegramEmailWithoutDisclaimer",
			TemplateData: map[string]interface{}{
				"Email": email,
			},
		}),
	)
	msg.ParseMode = "MarkdownV2"
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while editing a message!", zap.Error(err))
	}

	return nil
}
