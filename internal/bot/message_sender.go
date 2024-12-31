package bot

import (
	"errors"

	"gopkg.in/telebot.v4"
)

type MessageManager struct {
	bot    telebot.API
	origin *telebot.Message
}

func NewMessageManager(bot telebot.API, origin *telebot.Message) *MessageManager {
	return &MessageManager{
		bot:    bot,
		origin: origin,
	}
}

func (manager *MessageManager) Edit(content string) error {
	var err error
	manager.origin, err = manager.bot.Edit(manager.origin, content)
	if err != nil && !errors.Is(err, telebot.ErrMessageNotModified) {
		return err
	}
	return nil
}
