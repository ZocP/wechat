# Pickup Management System

A Go backend service for a WeChat Mini Program-based airport pickup management system, built with Gin, GORM, and Uber FX.

## Features

- **WeChat Mini Program Login** – One-click authentication via WeChat OAuth
- **Registration Management** – Passengers submit pickup requests with flight details
- **Order Management** – Create and track pickup orders with status workflow
- **WeChat Pay Integration** – Prepare payments and handle async notifications
- **Notice Board** – Publish and query announcements, filterable by flight number
- **Driver Assignment** – Assign drivers to orders and track acceptance
- **JWT Authentication** – Stateless auth with token refresh
- **Rate Limiting** – Per-IP request throttling middleware
- **OpenAPI Documentation** – Full API spec in `api/openapi.yaml`

## Tech Stack

| Component            | Technology         |
| -------------------- | ------------------ |
| Language             | Go 1.25+           |
| HTTP Framework       | Gin                |
| Database             | MySQL 8.0+         |
| ORM                  | GORM               |
| Dependency Injection | Uber FX            |
| Logging              | Zap + Lumberjack   |
| Configuration        | Viper              |
| Authentication       | JWT (golang-jwt/v5)|

## Project Structure

```
pickup/
├── api/                    # API documentation
│   └── openapi.yaml        # OpenAPI 3.0 specification
├── files/                  # Runtime files (generated)
│   ├── config.yaml         # Main configuration
│   └── logs/               # Log output
├── internal/               # Private application code
│   ├── config/             # Environment & DB configuration
│   ├── handler/            # HTTP request handlers
│   ├── middleware/          # Auth, rate-limit middleware
│   ├── model/              # GORM data models
│   ├── repository/         # Data access layer (DAL)
│   ├── service/            # Business logic layer
│   └── utils/              # JWT, crypto, WeChat utilities
├── pkg/                    # Reusable public packages
│   ├── config/             # Viper config loader
│   ├── server/             # Gin server bootstrap
│   └── zap/                # Zap logger setup
├── tests/                  # Handler-level tests with mocks
├── app.go                  # Application entry point
├── docker-compose.yml      # MySQL container setup
├── go.mod                  # Go module definition
└── env.example             # Environment variable template
```

## Quick Start

### Prerequisites

- Go 1.25+
- Docker (for MySQL) or a standalone MySQL 8.0+ instance
- WeChat Mini Program AppID & AppSecret (for production)

### 1. Start the Database

```bash
docker compose up -d
```

This launches a MySQL 8.0 container on port 3306 with database `pickup`.

### 2. Set Environment Variables

```bash
# Required
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=pickup
export DB_NAME=pickup
export JWT_SECRET=your_jwt_secret_key_should_be_long_and_random
export JWT_EXPIRE_HOURS=24
export JWT_ISSUER=pickup
export CRYPTO_KEY=your_crypto_key_32_characters_long

# WeChat (use test values for local development)
export WECHAT_APPID=your_wechat_appid
export WECHAT_SECRET=your_wechat_secret
```

See [env.example](env.example) for the full list including WeChat Pay settings.

### 3. Run the Server

```bash
go run app.go
```

The server starts at `http://localhost:8080` by default. To change the port, create `files/config.yaml`:

```yaml
server:
  port: 9090
  allowCORS: true
  releaseMode: false
```

### 4. Verify

```bash
curl http://localhost:8080/api/v1/health
# {"code":0,"message":"ok","data":null}
```

## API Endpoints

### Authentication

| Method | Path                        | Description              | Auth |
| ------ | --------------------------- | ------------------------ | ---- |
| POST   | `/api/v1/auth/wechat/login` | WeChat login             | No   |
| GET    | `/api/v1/auth/me`           | Get current user profile | Yes  |

### Registrations

| Method | Path                        | Description              | Auth |
| ------ | --------------------------- | ------------------------ | ---- |
| POST   | `/api/v1/registrations`     | Create a registration    | Yes  |
| GET    | `/api/v1/registrations`     | List my registrations    | Yes  |
| GET    | `/api/v1/registrations/my`  | List my registrations    | Yes  |
| GET    | `/api/v1/registrations/:id` | Get registration details | Yes  |
| PUT    | `/api/v1/registrations/:id` | Update a registration    | Yes  |
| DELETE | `/api/v1/registrations/:id` | Delete a registration    | Yes  |

### Orders

| Method | Path                              | Description                | Auth  |
| ------ | --------------------------------- | -------------------------- | ----- |
| POST   | `/api/v1/orders`                  | Create an order            | Yes   |
| GET    | `/api/v1/orders`                  | List my orders             | Yes   |
| GET    | `/api/v1/orders/:id`              | Get order details          | Yes   |
| POST   | `/api/v1/admin/orders/:id/notify` | Send order notification    | Admin |

### Payments

| Method | Path                  | Description             | Auth |
| ------ | --------------------- | ----------------------- | ---- |
| POST   | `/api/v1/pay/prepare` | Initiate WeChat payment | Yes  |
| POST   | `/api/v1/pay/notify`  | WeChat payment callback | No   |

### Notices

| Method | Path                                | Description           | Auth  |
| ------ | ----------------------------------- | --------------------- | ----- |
| GET    | `/api/v1/notices`                   | List visible notices  | Yes   |
| GET    | `/api/v1/notices/:id`               | Get notice details    | Yes   |
| GET    | `/api/v1/notices/flight/:flight_no` | Get notices by flight | Yes   |
| POST   | `/api/v1/admin/notices`             | Create a notice       | Admin |
| PUT    | `/api/v1/admin/notices/:id`         | Update a notice       | Admin |
| DELETE | `/api/v1/admin/notices/:id`         | Delete a notice       | Admin |

### Admin

| Method | Path                                    | Description            | Auth  |
| ------ | --------------------------------------- | ---------------------- | ----- |
| GET    | `/api/v1/admin/exports/database-fields` | Export DB schema fields | Admin |

## Testing

### Run All Tests with Coverage Report

```powershell
powershell -File run_tests.ps1
```

This script runs the full test suite, displays color-coded per-function coverage, and prints the overall total.

### Run All Tests (CLI)

```bash
go test ./... -v
```

### Run with Coverage Profile

```bash
go test ./... -coverprofile=coverage/coverage -count=1
go tool cover -func coverage/coverage
```

### Test Coverage Summary

| Package              | Coverage |
| -------------------- | -------- |
| internal/config      | 39.3%    |
| internal/handler     | 98.4%    |
| internal/middleware   | 86.2%    |
| internal/model       | 100.0%   |
| internal/repository  | 92.1%    |
| internal/service     | 85.3%    |
| internal/utils       | 89.4%    |
| pkg/config           | 83.3%    |
| pkg/server           | 63.8%    |
| pkg/zap              | 100.0%   |
| **Overall**          | **89.2%**|

Handler tests with mocks are in `tests/` (auth, order, payment, registration, notice). A Postman collection is available at [tests/postman_collection.json](tests/postman_collection.json).

## Configuration

### Environment Variables

| Variable            | Description                    | Default     |
| ------------------- | ------------------------------ | ----------- |
| `DB_HOST`           | MySQL host                     | `localhost` |
| `DB_PORT`           | MySQL port                     | `3306`      |
| `DB_USER`           | MySQL user                     | `root`      |
| `DB_PASSWORD`       | MySQL password                 | (empty)     |
| `DB_NAME`           | MySQL database name            | `pickup`    |
| `JWT_SECRET`        | JWT signing key                | (empty)     |
| `JWT_EXPIRE_HOURS`  | Token expiration in hours      | `24`        |
| `JWT_ISSUER`        | JWT issuer claim               | `pickup`    |
| `CRYPTO_KEY`        | AES encryption key (32 chars)  | (empty)     |
| `WECHAT_APPID`      | WeChat Mini Program App ID     | (empty)     |
| `WECHAT_SECRET`     | WeChat Mini Program App Secret | (empty)     |
| `WECHAT_MCH_ID`     | WeChat Pay Merchant ID         | (empty)     |
| `WECHAT_MCH_KEY`    | WeChat Pay Merchant Key        | (empty)     |
| `WECHAT_NOTIFY_URL` | WeChat Pay callback URL        | (empty)     |

### Server Configuration (`files/config.yaml`)

```yaml
server:
  addr: ""          # Bind address (empty = all interfaces)
  port: 8080        # Listen port
  allowCORS: true   # Enable CORS headers
  releaseMode: false # Gin release mode
```

## Deployment

### Docker Compose (Development)

```bash
docker compose up -d    # Start MySQL
go run app.go           # Start the API server
```

### Production Checklist

1. Set `server.releaseMode: true` in config
2. Use strong, random values for `JWT_SECRET` and `CRYPTO_KEY`
3. Configure real WeChat credentials and payment settings
4. Set up HTTPS via reverse proxy (Nginx, Caddy, etc.)
5. Configure log rotation (already handled by Lumberjack)
6. Set appropriate database connection pool sizes

## Development Guide

### Adding a New API Endpoint

1. Define the data model in `internal/model/`
2. Implement the repository in `internal/repository/`
3. Write business logic in `internal/service/`
4. Create the HTTP handler in `internal/handler/`
5. Register routes in `internal/handler/router.go`
6. Update `api/openapi.yaml`
7. Write unit tests

### Code Standards

- Format with `gofmt`
- Follow standard Go project layout conventions
- Keep functions small and single-purpose
- Use meaningful names; document exported symbols
- Write tests for all service-layer logic

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

