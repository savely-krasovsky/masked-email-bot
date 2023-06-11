package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (d *delivery) startCommand(update tgbotapi.Update) error {
	if err := d.service.StartCommand(update.Message.From.ID); err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Возникла ошибка! Повторите позже.")
		if _, err := d.bot.Send(msg); err != nil {
			d.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "Привет\\! Добавь токен \\(/token \\<token\\>\\) и отправь любую ссылку\\!")
	msg.ParseMode = "MarkdownV2"
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (d *delivery) tokenCommand(update tgbotapi.Update) error {
	if update.Message.CommandArguments() == "" {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Вы не отправили токен!")
		if _, err := d.bot.Send(msg); err != nil {
			d.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return nil
	}

	if err := d.service.TokenCommand(update.Message.From.ID, update.Message.CommandArguments()); err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Возникла ошибка! Повторите позже.")
		if _, err := d.bot.Send(msg); err != nil {
			d.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "Токен успешно сохранён!")
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (d *delivery) authCommand(update tgbotapi.Update) error {
	authCodeURL, err := d.service.AuthCommand(update.Message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Возникла ошибка! Повторите позже.")
		if _, err := d.bot.Send(msg); err != nil {
			d.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "Please sign in using your Fastmail account.")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
		{
			Text: "Sign in with Fastmail",
			URL:  &authCodeURL,
		},
	})

	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (d *delivery) anyOtherCommand(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.From.ID, "Такой команды нет!")
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (d *delivery) link(update tgbotapi.Update) error {
	email, err := d.service.Link(update.Message.From.ID, update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Возникла ошибка! Повторите позже.")
		if _, err := d.bot.Send(msg); err != nil {
			d.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "`"+email+"`")
	msg.ParseMode = "MarkdownV2"
	if _, err := d.bot.Send(msg); err != nil {
		d.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}
