package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

type Bot struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func (b Bot) sendMessage(message string) error {
	msg := tgbotapi.NewMessage(b.chatID, message)
	msg.ParseMode = "HTML"
	_, err := b.bot.Send(msg)
	return err
}

func (b Bot) Notify(message string) error {
	return b.sendMessage(message)
}

func New(token string, chatID int64) (Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return Bot{}, fmt.Errorf("create bot error: %w", err)
	}
	return Bot{bot: bot, chatID: chatID}, err
}

func NewFromViper(v *viper.Viper) (Bot, error) {
	token := v.GetString("token")
	if token == "" {
		return Bot{}, fmt.Errorf("token is required")
	}
	chatID := v.GetInt64("chat_id")
	if chatID == 0 {
		return Bot{}, fmt.Errorf("chat_id is required")
	}
	return New(token, chatID)
}
