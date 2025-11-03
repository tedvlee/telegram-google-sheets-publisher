package publisher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type InlineKeyboardButton struct {
	Text string `json:"text"`
	URL  string `json:"url,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type TelegramMessage struct {
	ChatID        string                `json:"chat_id"`
	Text          string                `json:"text"`
	ParseMode     string                `json:"parse_mode,omitempty"`
	ReplyMarkup   *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	DisableWebPagePreview bool           `json:"disable_web_page_preview,omitempty"`
}

func SendTelegramMessage(token, chatID, text string, buttons [][2]string) (string, error) {
	var markup *InlineKeyboardMarkup
	if len(buttons) > 0 {
		var row []InlineKeyboardButton
		for _, btn := range buttons {
			if btn[0] != "" && btn[1] != "" {
				row = append(row, InlineKeyboardButton{Text: btn[0], URL: btn[1]})
			}
		}
		if len(row) > 0 {
			markup = &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{row}}
		}
	}

	msg := TelegramMessage{
		ChatID:        chatID,
		Text:          text,
		ParseMode:     "HTML",
		ReplyMarkup:   markup,
		DisableWebPagePreview: true,
	}

	body, _ := json.Marshal(msg)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Ok          bool `json:"ok"`
		Description string `json:"description,omitempty"`
		Result      struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if !result.Ok {
		return "", fmt.Errorf("Telegram error: %s", result.Description)
	}
	return fmt.Sprintf("%d", result.Result.MessageID), nil
}