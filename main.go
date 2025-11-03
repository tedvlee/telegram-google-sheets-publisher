package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"telegram-publisher/publisher"
)

func main() {
	godotenv.Load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	credsFile := "credentials.json"

	if token == "" || chatID == "" || spreadsheetID == "" {
		slog.Error("Missing env vars")
		os.Exit(1)
	}

	// Запуск сразу при старте
	publishPosts(token, chatID, spreadsheetID, credsFile)

	// Запуск по расписанию (каждый час)
	c := cron.New()
	c.AddFunc("@hourly", func() {
		publishPosts(token, chatID, spreadsheetID, credsFile)
	})
	c.Start()

	// Keep alive
	select {}
}

func publishPosts(token, chatID, spreadsheetID, credsFile string) {
	ctx := context.Background()
	posts, err := publisher.ReadPendingPosts(ctx, spreadsheetID, credsFile)
	if err != nil {
		slog.Error("Failed to read posts", "error", err)
		return
	}

	for _, post := range posts {
		slog.Info("Publishing post", "row", post.RowIndex)

		buttons := [][2]string{}
		if post.Btn1Text != "" && post.Btn1URL != "" {
			buttons = append(buttons, [2]string{post.Btn1Text, post.Btn1URL})
		}
		if post.Btn2Text != "" && post.Btn2URL != "" {
			buttons = append(buttons, [2]string{post.Btn2Text, post.Btn2URL})
		}

		msgID, err := publisher.SendTelegramMessage(token, chatID, post.Text, buttons)
		status := "✅ Опубликовано"
		if err != nil {
			status = "❌ Ошибка: " + err.Error()
			msgID = ""
			slog.Error("Failed to send", "error", err)
		}

		publisher.UpdateStatus(ctx, spreadsheetID, credsFile, post.RowIndex, status, msgID)
		time.Sleep(1 * time.Second) // уважаем Telegram
	}
}