package telegram

import (
	"errors"
	"regexp"
	"strings"

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

func (d *delivery) generateMaskedEmail(localizer *i18n.Localizer, update tgbotapi.Update) error {
	maskedEmail, err := d.service.GenerateMaskedEmail(update.Message.From.ID, update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "TelegramError",
		}))
		if _, err := d.bot.Send(msg); err != nil {
			d.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}
	maskedEmail.ID = "id:" + maskedEmail.ID

	msg := tgbotapi.NewMessage(update.Message.From.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "TelegramEmail",
		TemplateData: map[string]interface{}{
			"Email": maskedEmail.Email,
		},
	}))
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{{
		Text:         localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "TelegramEmailDoNotDeleteButton"}),
		CallbackData: &maskedEmail.ID,
	}})
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (d *delivery) enableMaskedEmail(localizer *i18n.Localizer, update tgbotapi.Update) error {
	dataParts := strings.Split(update.CallbackData(), ":")
	if len(dataParts) < 2 {
		return errors.New("invalid callback data")
	}
	id := dataParts[1]

	if err := d.service.EnableMaskedEmail(update.CallbackQuery.From.ID, id); err != nil {
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

	// Remove inline keyboard and remove disclaimer from message
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

func (d *delivery) answerInlineQueryWithEmail(localizer *i18n.Localizer, update tgbotapi.Update) error {
	re := regexp.MustCompile(`[a-z0-9_]+`)
	if re.FindString(update.InlineQuery.Query) != update.InlineQuery.Query || update.InlineQuery.Query == "" {
		inlineConf := tgbotapi.InlineConfig{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			CacheTime:     0,
			Results:       []interface{}{},
		}
		if _, err := d.bot.Request(inlineConf); err != nil {
			d.logger.Error("Error while answering inline query!", zap.Error(err))
		}

		return nil
	}

	example := update.InlineQuery.Query + ".xxxxx@example.com"
	result := tgbotapi.NewInlineQueryResultArticleMarkdownV2(update.InlineQuery.ID, example, "`"+example+"`")
	update.InlineQuery.Query = "prefix:" + update.InlineQuery.Query
	markup := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{{
		Text:         localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "TelegramInlineQueryGenerate"}),
		CallbackData: &update.InlineQuery.Query,
	}})
	result.ReplyMarkup = &markup

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{result},
	}
	if _, err := d.bot.Request(inlineConf); err != nil {
		d.logger.Error("Error while answering inline query!", zap.Error(err))
	}

	return nil
}

func (d *delivery) generateMaskedEmailWithInlineButton(localizer *i18n.Localizer, update tgbotapi.Update) error {
	maskedEmail, err := d.service.Prefix(update.CallbackQuery.From.ID, strings.Split(update.CallbackData(), ":")[1])
	if err != nil {
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
		MessageID: "TelegramInlineQueryGenerated",
	}))
	if _, err := d.bot.Request(callback); err != nil {
		d.logger.Error("Error while answering to the callback query!", zap.Error(err))
	}

	msg := &tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			InlineMessageID: update.CallbackQuery.InlineMessageID,
		},
		Text: "`" + maskedEmail.Email + "`",
	}
	msg.ParseMode = "MarkdownV2"
	if _, err := d.bot.Request(msg); err != nil {
		d.logger.Error("Error while editing a message!", zap.Error(err))
	}

	return nil
}
