# dp-vc-webApp

A production-oriented REST API backend built with Go.

## Architecture

This project follows a clean architecture pattern with clear separation of concerns:

```
dp-vc-webApp/
├── main.go                 # Application entry point and bootstrap
├── go.mod                  # Go module definition
├── go.sum                  # Dependency lock file
├── Makefile                # Build and development commands
├── Dockerfile              # Container image definition
├── docker-compose.yml      # Local infrastructure (MongoDB, Redis)
├── .env.example            # Environment configuration template
│
├── assets/                 # Static assets (images, files)
├── common/                 # Shared utilities and helpers
├── configs/                # Configuration and infrastructure
│   ├── auth/              # Authentication middleware and helpers
│   ├── env/               # Environment configuration
│   ├── mongo/             # MongoDB connection and operations
│   ├── observability/     # OpenTelemetry tracing
│   ├── redis/             # Redis connection and caching
│   ├── response/          # Response formatting helpers
│   └── types/             # Shared type definitions
│
├── controllers/            # HTTP request handlers
│   └── health/           # Health check endpoints
│
├── models/                 # Data structures and MongoDB schemas
│   └── health/           # Health models
│
├── routes/                 # Route definitions
│   ├── health.go         # Health routes
│   └── routes.go         # Route registration
│
├── services/              # Business logic layer
├── libraries/             # Reusable libraries
├── utils/                 # Utility functions
├── webhooks/              # Webhook handlers
│
├── worker/                # Background worker
│   ├── workers.go       # Worker implementation
│   └── jobs/            # Job definitions
│
├── docs/                   # API documentation
├── scripts/                # Automation scripts
├── schemas/                # JSON schemas
├── dbs/                    # Database migration/seed data
├── environments/           # Environment-specific configs
├── logs/                   # Log files (gitignored)
└── tmp/                    # Temporary files (gitignored)
```

## Directory Responsibilities

| Package | Purpose |
|---------|---------|
| `configs/` | Configuration, infrastructure clients, middleware |
| `controllers/` | HTTP request handling, validation, calling services |
| `models/` | Data structures, MongoDB document definitions |
| `routes/` | Route definitions, middleware attachment |
| `services/` | Business logic and workflows |
| `worker/` | Background job processing |
| `utils/` | General utility functions |

## Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Make (optional)

## Environment Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Configure environment variables:
   ```env
   PORT=8080

   MONGO_URI=mongodb://localhost:27017
   MONGO_DB_NAME=dp-vc-webApp

   REDIS_URI=localhost:6379
   REDIS_PASSWORD=
   REDIS_USER=
   REDIS_PREFIX=dp-vc-webApp

   OTEL_EXPORTER_OTLP_ENDPOINT=
   APP_ENV=development
   ```

## Quick Start

```bash
# Install dependencies
make install

# Start infrastructure (MongoDB, Redis)
make docker-up

# Run the server
make run

# In another terminal, test the endpoint
curl http://localhost:8080/health
```

`docker-compose.yml` starts three application services: `api`, `worker`, and the local infrastructure. The worker runs continuously with:

```bash
./server --worker
```

It scans for payment notifications every 30 minutes by default. Start everything with:

```bash
docker compose up -d --build
```

View worker logs with:

```bash
docker compose logs -f worker
```

## Deploying on Replit

1. Import this repository into a new Replit workspace.
2. In **Tools > Secrets**, add the variables from `.env.example`. At minimum, set `MONGO_URI`, `MONGO_DB_NAME`, and a strong `AUTH_SECRET`.
3. Click **Run**. The included `.replit` file starts the API with `go run .`; the application reads Replit's `PORT` variable automatically.
4. Verify the deployment at `/health`.

MongoDB must be hosted externally, such as MongoDB Atlas. Redis is optional for the current payment/auth API; if it is unavailable, the application logs a warning and continues. Do not upload `.env` or place passwords, SMTP credentials, Telegram tokens, or database credentials in source code.

The payment reminder worker is a separate process. Run `go run main.go --worker` in a second always-on Replit deployment or scheduled worker. A normal web deployment only serves HTTP requests.

## Running the Server

```bash
# Using Make
make run

# Or directly
go run main.go
```

The server will start on `http://0.0.0.0:8080`

## Running as Worker

```bash
# Using Make
make worker

# Or directly
go run main.go --worker
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test -v -race -cover ./...

# Run specific package tests
go test -v ./controllers/health/...
```

## Docker Commands

```bash
# Start infrastructure
make docker-up

# View logs
make docker-logs

# Stop infrastructure
make docker-down

# Reset (remove volumes)
make docker-reset
```

## Health Endpoint

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "success",
  "data": {
    "service": "dp-vc-webApp",
    "status": "healthy"
  }
}
```

## Payment Notifications

Run the worker separately with `go run main.go --worker`. By default, it scans the `payments` collection every 30 minutes. It sends reminders for pending payments due within `PAYMENT_REMINDER_DAYS`, including overdue pending payments, using a Redis lock so only one worker sends a reminder when multiple worker instances are running. MongoDB remains the source of payment records; Redis is only used for the distributed scheduler lock.

For a system cron job that starts and exits after one scan, use:

```cron
*/30 * * * * cd /path/to/backend && /usr/local/go/bin/go run . --notifications-once >> logs/notifications.log 2>&1
```

For production, build the binary once and use the binary in cron instead of `go run`:

```bash
go build -o bin/server .
```

```cron
*/30 * * * * cd /path/to/backend && ./bin/server --notifications-once >> logs/notifications.log 2>&1
```

For Gmail, create a Gmail App Password and configure:

```env
GMAIL_USERNAME=your-gmail@gmail.com
GMAIL_APP_PASSWORD=your-16-character-app-password
GMAIL_FROM=your-gmail@gmail.com
```

The server sends through `smtp.gmail.com:587`. Never use your normal Gmail password and never expose these values to the frontend.

For Telegram, create a bot with BotFather, set `TELEGRAM_BOT_TOKEN`, and add each recipient's chat ID to the payment request. Each recipient must start or interact with the bot before the bot can message them.

## Authentication API

Create an account with `POST /auth/signup` or sign in with `POST /auth/login`. Both return a bearer token. Send it with every payment request:

```http
Authorization: Bearer <token>
```

Signup body:

```json
{"name":"A User","email":"user@example.com","password":"at-least-8-characters"}
```

Login body:

```json
{"email":"user@example.com","password":"at-least-8-characters"}
```

### Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `POST` | `/payments` | Create a pending payment |
| `GET` | `/payments?status=pending` | List payments |
| `GET` | `/payments/:id` | Get one payment |
| `PATCH` | `/payments/:id/status` | Mark `pending`, `paid`, or `cancelled` |
| `DELETE` | `/payments/:id` | Delete a payment |

Example create body:

```json
{
   "title": "Hosting invoice",
   "description": "Monthly production hosting",
   "amount": 49.99,
   "currency": "USD",
   "due_date": "2026-10-01T09:00:00Z",
   "recipient_name": "Finance team",
   "recipient_emails": ["finance@example.com", "owner@example.com"],
   "telegram_chat_ids": ["123456789", "987654321"],
   "notification_channels": ["email"]
}
```

### Frontend integration prompt

> Build a payment-tracking screen connected to `http://localhost:8080`. Authenticate with `POST /auth/signup` or `POST /auth/login`, save the returned bearer token securely, and send it as `Authorization: Bearer <token>` for payment requests. Provide a form for title, description, amount, three-letter currency, UTC due date/time, recipient name, multiple recipient emails, multiple Telegram chat IDs, and notification channels (`email` and/or `telegram`). Submit with `POST /payments`; list records with `GET /payments?status=pending`; show due date, amount, recipients, channels, and status; allow marking a record paid or cancelled with `PATCH /payments/:id/status` and deleting it with `DELETE /payments/:id`. Use the API envelope `{ status, data, error }`, show validation/server errors, display dates in the user’s local timezone, and refresh the list after every mutation. Do not put Gmail, SMTP, Redis, MongoDB, or Telegram secrets in the browser; those remain server-side environment variables.

## Adding New Features

To add a new domain (e.g., `users`):

1. **Create controller:**
   ```
   controllers/users/users.go
   ```

2. **Create models:**
   ```
   models/users/user.go
   ```

3. **Create routes:**
   ```go
   // routes/users.go
   func Users(router *gin.Engine) {
       group := router.Group("/users")
       group.Use(auth.Auth(types.RouterArr{
           "/users": {"users.view"},
       }))
       group.GET("", usersController.List)
   }
   ```

4. **Register routes:**
   ```go
   // routes/routes.go
   func Register(router *gin.Engine) {
       Health(router)
       Users(router) // Add this line
   }
   ```

5. **Create service (optional):**
   ```go
   // services/users/users.go
   func GetAll(ctx context.Context) ([]models.User, error) { ... }
   ```

## Available Make Commands

| Command | Description |
|---------|-------------|
| `make run` | Start HTTP server |
| `make worker` | Start background worker |
| `make test` | Run tests |
| `make fmt` | Format code |
| `make vet` | Run go vet |
| `make lint` | Run golangci-lint |
| `make build` | Build binary |
| `make docker-up` | Start Docker containers |
| `make docker-down` | Stop Docker containers |
| `make help` | Show all commands |

## Technology Stack

- **Language:** Go 1.23+
- **Router:** Gin
- **Database:** MongoDB
- **Cache:** Redis
- **Tracing:** OpenTelemetry
- **Container:** Docker

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Submit a pull request
