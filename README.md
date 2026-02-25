# pantry-api

A Go REST API for managing pantry items and locations, with expiry tracking and notifications.

## Features

- CRUD for pantry locations and items
- Filtering items by tags and location
- Expiry tracking with configurable lifespan
- Expiry notifications via Infobip (email), Telegram, or terminal
- Firebase authentication
- Firestore as the database
- OpenTelemetry tracing

## Interfaces

There is SPA Web UI available at [pantry-web](https://github.com/nickelghost/pantry-web).

## Motivations

This project was created to solve household issues with expiring items. I've used it to practice Go, REST API design and cloud services integration. It is not intended for production use and may have security and performance issues. Use at your own risk.

## Modes

The application runs in one of two modes, controlled by the `MODE` environment variable:

| `MODE`       | Description                                                   |
| ------------ | ------------------------------------------------------------- |
| _(unset)_    | Starts the HTTP API server on port `8080`                     |
| `notify_job` | Runs a one-shot job that sends expiry notifications and exits |

## API

| Method   | Path                   | Description                          |
| -------- | ---------------------- | ------------------------------------ |
| `GET`    | `/locations`           | List all locations with their items  |
| `GET`    | `/locations/{id}`      | Get a single location with its items |
| `POST`   | `/locations`           | Create a location                    |
| `PUT`    | `/locations/{id}`      | Update a location                    |
| `DELETE` | `/locations/{id}`      | Delete a location                    |
| `POST`   | `/items`               | Create an item                       |
| `PUT`    | `/items/{id}`          | Update an item                       |
| `PATCH`  | `/items/{id}/location` | Update an item's location            |
| `DELETE` | `/items/{id}`          | Delete an item                       |
| `GET`    | `/healthz`             | Health check                         |

Both `/locations` and `/locations/{id}` accept an optional `tags` query parameter (comma-separated) to filter items.

## Environment Variables

### API server

| Variable                       | Description                                                |
| ------------------------------ | ---------------------------------------------------------- |
| `ACCESS_CONTROL_ALLOW_ORIGIN`  | Comma-separated list of allowed CORS origins               |
| `ACCESS_CONTROL_ALLOW_HEADERS` | Comma-separated list of allowed CORS headers               |
| `FIREBASE_AUTH_DISABLED`       | Set to `true` to disable authentication (development only) |

### Notifications (`notify_job`)

Exactly one notifier should be configured:

**Infobip (email)**

| Variable               | Description          |
| ---------------------- | -------------------- |
| `INFOBIP_API_BASE_URL` | Infobip API base URL |
| `INFOBIP_API_KEY`      | Infobip API key      |
| `INFOBIP_FROM`         | Sender email address |

**Telegram**

| Variable           | Description                               |
| ------------------ | ----------------------------------------- |
| `TELEGRAM_TOKEN`   | Telegram bot token                        |
| `TELEGRAM_CHAT_ID` | Telegram chat ID to send notifications to |

If neither is configured, notifications are printed to the terminal.

### Optional

| Variable                         | Description                                                                     |
| -------------------------------- | ------------------------------------------------------------------------------- |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to GCP service account JSON (if not using Application Default Credentials) |
| `LOG_FORMAT`                     | Log format (`json` or unset for text)                                           |
| `LOG_LEVEL`                      | Log level (`debug`, `info`, `warn`, `error`)                                    |

## Development

**Prerequisites:** Go 1.24+, [Mage](https://magefile.org/)

```sh
# Run tests with coverage
mage test

# Build
go build -o bin/pantry-api .

# Run locally (requires a .env file)
go run .
```

A `.env` file is loaded automatically via `godotenv`.

## Docker

```sh
docker build -t pantry-api .
docker run -p 8080:8080 --env-file .env pantry-api
```

The image is built from `scratch` with CA certificates included for TLS.
