package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var errTelegramAPI = errors.New("telegram API error")

type telegramNotifier struct {
	client *http.Client
	token  string
	chatID int
}

func (n telegramNotifier) getURL(subpath string) url.URL {
	return url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   "/bot" + n.token + subpath,
	}
}

func (n telegramNotifier) NotifyAboutItems(
	ctx context.Context,
	expiries []itemExpiry,
	comingExpiries []itemExpiry,
	_ authenticationRepository,
) error {
	msg := getNotificationTitle() + "\n\n" + notificationExpiriesToText(expiries, comingExpiries)
	targetURL := n.getURL("/sendMessage")

	body, err := json.Marshal(map[string]any{
		"chat_id": n.chatID,
		"text":    msg,
	})
	if err != nil {
		return fmt.Errorf("marshal api request body json: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request to telegram api: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", errTelegramAPI, res.Status)
	}

	return nil
}
