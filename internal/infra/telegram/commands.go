package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (a *adapter) startCommand(update tgbotapi.Update) error {
	if err := a.service.StartCommand(int(update.Message.From.ID)); err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Возникла ошибка! Повторите позже.")
		if _, err := a.bot.Send(msg); err != nil {
			a.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "Привет\\! Добавь токен \\(/token \\<token\\>\\) и отправь любую ссылку\\!")
	msg.ParseMode = "MarkdownV2"
	if _, err := a.bot.Send(msg); err != nil {
		a.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (a *adapter) tokenCommand(update tgbotapi.Update) error {
	if update.Message.CommandArguments() == "" {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Вы не отправили токен!")
		if _, err := a.bot.Send(msg); err != nil {
			a.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return nil
	}

	if err := a.service.TokenCommand(int(update.Message.From.ID), update.Message.CommandArguments()); err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Возникла ошибка! Повторите позже.")
		if _, err := a.bot.Send(msg); err != nil {
			a.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "Токен успешно сохранён!")
	if _, err := a.bot.Send(msg); err != nil {
		a.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (a *adapter) anyOtherCommand(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.From.ID, "Такой команды нет!")
	if _, err := a.bot.Send(msg); err != nil {
		a.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}

func (a *adapter) link(update tgbotapi.Update) error {
	email, err := a.service.Link(int(update.Message.From.ID), update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "Возникла ошибка! Повторите позже.")
		if _, err := a.bot.Send(msg); err != nil {
			a.logger.Error("Error while sending a message!", zap.Error(err))
		}
		return err
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "`"+email+"`")
	msg.ParseMode = "MarkdownV2"
	if _, err := a.bot.Send(msg); err != nil {
		a.logger.Error("Error while sending a message!", zap.Error(err))
	}

	return nil
}
