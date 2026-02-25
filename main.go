package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nickelghost/nglog"
	"github.com/nickelghost/ngtel"
)

const (
	errExitCode      = 1
	otelFailExitCode = 2
)

var errOtelConfigFail = errors.New("failed configuring otel")

func main() {
	ctx := context.Background()

	nglog.SetUpLogger(os.Stderr, os.Getenv("LOG_FORMAT"), nglog.GetLogLevel(os.Getenv("LOG_LEVEL")))

	err := start(ctx)
	switch {
	case errors.Is(err, errOtelConfigFail):
		slog.Error("failed configuring tracing", "err", err)
		os.Exit(otelFailExitCode)
	case err != nil:
		slog.Error("failed to start", "err", err)
		os.Exit(errExitCode)
	}
}

func start(ctx context.Context) error {
	tracerShutdown, err := ngtel.ConfigureOtel(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errOtelConfigFail, err)
	}

	defer tracerShutdown()

	switch strings.ToLower(os.Getenv("MODE")) {
	case "notify_job":
		err = initNotifyJob(ctx)
	default:
		err = initAPI(ctx)
	}

	return err
}

func initAPI(ctx context.Context) error {
	validate := getValidate()

	firestoreRepo, err := getFirestoreRepository(ctx)
	if err != nil {
		return err
	}

	defer firestoreRepo.client.Close() //nolint:errcheck

	auth, err := getFirebaseAuthentication(ctx)
	if err != nil {
		return err
	}

	srv := getServer(getRouter(firestoreRepo, validate, auth))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func initNotifyJob(ctx context.Context) error {
	httpClient := &http.Client{Timeout: httpTimeout}

	firestoreRepo, err := getFirestoreRepository(ctx)
	if err != nil {
		return err
	}

	defer firestoreRepo.client.Close() //nolint:errcheck

	auth, err := getFirebaseAuthentication(ctx)
	if err != nil {
		return err
	}

	authRepo := firebaseAuthenticationRepository{client: auth.client}

	var n notifier

	if baseURL := os.Getenv("INFOBIP_API_BASE_URL"); baseURL != "" {
		n = infobipNotifier{
			client:  httpClient,
			baseURL: baseURL,
			apiKey:  os.Getenv("INFOBIP_API_KEY"),
			from:    os.Getenv("INFOBIP_FROM"),
		}
	} else if token := os.Getenv("TELEGRAM_TOKEN"); token != "" {
		chatID, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
		if err != nil {
			return fmt.Errorf("invalid Telegram chat id: %w", err)
		}

		n = telegramNotifier{client: httpClient, token: token, chatID: chatID}
	} else {
		n = terminalNotifier{}
	}

	if err := notifyAboutItems(ctx, firestoreRepo, n, authRepo); err != nil {
		slog.Error("failed to notify about items", "err", err)
	}

	return nil
}

func getValidate() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}
