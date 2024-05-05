package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

var errInfobipAPI = errors.New("infobip API error")

type infobipNotifier struct {
	client  *http.Client
	baseURL string
	apiKey  string
	from    string
}

func (n infobipNotifier) getURL() url.URL {
	return url.URL{
		Scheme: "https",
		Host:   n.baseURL,
	}
}

func (n infobipNotifier) getRequest(
	ctx context.Context, method string, url string, payload io.Reader, contentType string,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, payload)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "App "+n.apiKey)
	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (n infobipNotifier) getEmailPayload(subject string, text string, emails []string) (io.Reader, string, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("from", n.from)
	_ = writer.WriteField("subject", subject)
	_ = writer.WriteField("text", text)

	for _, email := range emails {
		_ = writer.WriteField("to", email)
	}

	err := writer.Close()
	if err != nil {
		return nil, "", fmt.Errorf("close payload writer: %w", err)
	}

	return payload, writer.FormDataContentType(), nil
}

func (n infobipNotifier) NotifyAboutItems(
	ctx context.Context,
	expiries []itemExpiry,
	comingExpiries []itemExpiry,
	authRepo authenticationRepository,
) error {
	emails, err := authRepo.GetAllEmails(ctx)
	if err != nil {
		return fmt.Errorf("get all emails: %w", err)
	}

	infobipURL := n.getURL()
	infobipURL.Path = "/email/3/send"

	payload, contentType, err := n.getEmailPayload(
		getNotificationTitle(), notificationExpiriesToText(expiries, comingExpiries), emails,
	)
	if err != nil {
		return err
	}

	req, err := n.getRequest(ctx, "POST", infobipURL.String(), payload, contentType)
	if err != nil {
		return err
	}

	res, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("do http request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %w", res.Status, errInfobipAPI)
	}

	return nil
}
