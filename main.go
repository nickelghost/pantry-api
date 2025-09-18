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
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/nickelghost/nglog"
	"github.com/nickelghost/ngtel"
	"github.com/nickelghost/ngtelgcp"
	"go.opentelemetry.io/otel/trace"
)

const httpTimeout = 10 * time.Second

func getValidate() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

func getFirebaseAuthentication(ctx context.Context, tracer trace.Tracer) (firebaseAuthentication, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: os.Getenv("CLOUDSDK_CORE_PROJECT"),
	})
	if err != nil {
		return firebaseAuthentication{}, fmt.Errorf("failed to create firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return firebaseAuthentication{}, fmt.Errorf("failed to create firebase auth client: %w", err)
	}

	return firebaseAuthentication{client: client, tracer: tracer}, nil
}

func getFirestoreRepository(ctx context.Context, tracer trace.Tracer) (firestoreRepository, error) {
	client, err := firestore.NewClientWithDatabase(ctx,
		os.Getenv("CLOUDSDK_CORE_PROJECT"),
		os.Getenv("FIRESTORE_DATABASE"),
	)
	if err != nil {
		return firestoreRepository{}, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return firestoreRepository{client: client, tracer: tracer}, nil
}

func initNotifyJob(ctx context.Context) error {
	tpOpts, resOpts, err := ngtelgcp.GetTracerOpts()
	if err != nil {
		return err
	}

	tracer, tracerShutdown, err := ngtel.CreateTracer(
		ctx, os.Getenv("ENABLE_TRACING") == "true", "pantry-api", tpOpts, resOpts,
	)
	if err != nil {
		return err
	}

	defer tracerShutdown()

	httpClient := &http.Client{Timeout: httpTimeout}

	firestoreRepo, err := getFirestoreRepository(ctx, tracer)
	if err != nil {
		return err
	}

	defer firestoreRepo.client.Close()

	auth, err := getFirebaseAuthentication(ctx, tracer)
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

func initAPI(ctx context.Context) error {
	validate := getValidate()

	tpOpts, resOpts, err := ngtelgcp.GetTracerOpts()
	if err != nil {
		return err
	}

	tracer, tracerShutdown, err := ngtel.CreateTracer(
		ctx, os.Getenv("ENABLE_TRACING") == "true", "pantry-api", tpOpts, resOpts,
	)
	if err != nil {
		return err
	}

	defer tracerShutdown()

	firestoreRepo, err := getFirestoreRepository(ctx, tracer)
	if err != nil {
		return err
	}

	defer firestoreRepo.client.Close()

	auth, err := getFirebaseAuthentication(ctx, tracer)
	if err != nil {
		return err
	}

	srv := getServer(getRouter(firestoreRepo, validate, auth))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "err", err)
	}

	return nil
}

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	nglog.SetUpLogger(os.Stderr, os.Getenv("LOG_FORMAT"), nglog.GetLogLevel(os.Getenv("LOG_LEVEL")))

	var err error

	switch strings.ToLower(os.Getenv("MODE")) {
	case "notify_job":
		err = initNotifyJob(ctx)
	default:
		err = initAPI(ctx)
	}

	if err != nil {
		slog.Error("failed to initialize", "err", err)
		os.Exit(1)
	}
}
