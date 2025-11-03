package publisher

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

type PostRow struct {
	RowIndex       int
	Status         string
	PublishTimeStr string
	Text           string
	Btn1Text       string
	Btn1URL        string
	Btn2Text       string
	Btn2URL        string
}

func ReadPendingPosts(ctx context.Context, spreadsheetID string, credsFile string) ([]PostRow, error) {
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		return nil, err
	}

	rangeName := "Посты!A2:H1000" // A=Status, B=Time, C=Text, D-G=Buttons, H=MsgID
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeName).Do()
	if err != nil {
		return nil, err
	}

	var posts []PostRow
	for i, row := range resp.Values {
		if len(row) < 3 {
			continue
		}

		status := row[0].(string)
		if status == "✅ Опубликовано" {
			continue
		}

		var publishTimeStr string
		if len(row) > 1 && row[1] != nil {
			publishTimeStr = row[1].(string)
			if publishTimeStr != "" {
				// Попытка распарсить дату
				if t, err := time.Parse("02.01.2006 15:04", publishTimeStr); err == nil {
					if time.Now().Before(t) {
						continue // ещё не время
					}
				}
			}
		}

		text := row[2].(string)
		if text == "" {
			continue
		}

		var btn1Text, btn1URL, btn2Text, btn2URL string
		if len(row) > 3 && row[3] != nil {
			btn1Text = row[3].(string)
		}
		if len(row) > 4 && row[4] != nil {
			btn1URL = row[4].(string)
		}
		if len(row) > 5 && row[5] != nil {
			btn2Text = row[5].(string)
		}
		if len(row) > 6 && row[6] != nil {
			btn2URL = row[6].(string)
		}

		posts = append(posts, PostRow{
			RowIndex:       i + 2, // Google Sheets rows start at 1, +1 for header
			Status:         status,
			PublishTimeStr: publishTimeStr,
			Text:           text,
			Btn1Text:       btn1Text,
			Btn1URL:        btn1URL,
			Btn2Text:       btn2Text,
			Btn2URL:        btn2URL,
		})
	}
	return posts, nil
}

func UpdateStatus(ctx context.Context, spreadsheetID, credsFile string, rowIdx int, status string, msgID string) error {
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		return err
	}

	// Обновить столбец A (статус) и H (ID сообщения)
	statusRange := "Посты!A" + string(rune('0'+rowIdx))
	msgIDRange := "Посты!H" + string(rune('0'+rowIdx))

	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, statusRange, &sheets.ValueRange{
		Values: [][]interface{}{{status}},
	}).ValueInputOption("RAW").Do()
	if err != nil {
		slog.Error("Failed to update status", "error", err)
	}

	if msgID != "" {
		_, err = srv.Spreadsheets.Values.Update(spreadsheetID, msgIDRange, &sheets.ValueRange{
			Values: [][]interface{}{{msgID}},
		}).ValueInputOption("RAW").Do()
		if err != nil {
			slog.Error("Failed to update message ID", "error", err)
		}
	}
	return nil
}