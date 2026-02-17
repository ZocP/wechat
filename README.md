# UIUC Pickup Scheduler Backend

Go backend for the UIUC international student airport pickup scheduling workflow.

## Current Scope

- WeChat login (`code -> open_id`) and JWT issuance
- Phone binding via WeChat `getuserphonenumber`
- Student pickup request submission/update
- Admin shift/driver management and transactional assignment
- Shift publish workflow
- Flight sync job placeholder (configurable, safe-noop when not configured)

## Tech Stack

- Go + Gin
- GORM + MySQL 8.0
- Uber Fx
- JWT auth
- robfig/cron v3

## Active API Surface

Base path: `/api/v1`

- `GET /health`
- `POST /auth/login`
- `POST /auth/bind-phone` (JWT)
- `POST /student/requests` (student)
- `GET /student/requests/my` (student)
- `PUT /student/requests/:id` (student)
- `GET /admin/drivers` (admin)
- `POST /admin/drivers` (admin)
- `GET /admin/shifts/dashboard` (admin)
- `GET /admin/requests/pending` (admin)
- `POST /admin/shifts` (admin)
- `POST /admin/shifts/:id/assign-student` (admin)
- `POST /admin/shifts/:id/remove-student` (admin)
- `POST /admin/shifts/:id/assign-staff` (admin)
- `POST /admin/shifts/:id/publish` (admin)

OpenAPI source: [api/openapi.yaml](api/openapi.yaml)

## Quick Start

1) Start MySQL (or use your own MySQL 8.0 instance)

```bash
docker compose up -d
```

2) Configure env

```bash
cp env.example .env
```

Then export/load variables from `.env` (or set them directly in your shell).

3) Optional server config file

- Copy [files/config.template.yaml](files/config.template.yaml) to `files/config.yaml`
- Adjust server address/port/CORS/release mode as needed

4) Run

```bash
go run app.go
```

## Configuration

### Environment Variables

See [env.example](env.example). Important keys:

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET`, `JWT_EXPIRE_HOURS`, `JWT_ISSUER`
- `WECHAT_APPID`, `WECHAT_SECRET`
- `WECHAT_MCH_ID`, `WECHAT_MCH_KEY`, `WECHAT_NOTIFY_URL`
- `CRYPTO_KEY`
- `FLIGHT_API_URL` (optional; cron sync skips when empty)

### File-based Config

Template: [files/config.template.yaml](files/config.template.yaml)

Runtime file (optional): `files/config.yaml`

## Testing

```bash
go test ./...
```

Coverage example:

```bash
go test ./... -coverprofile=coverage_all -covermode=atomic
go tool cover -func coverage_all
```

## Notes

- This repository still contains some legacy modules from the earlier registration/order/payment flow.
- The scheduler routes above are the active main flow wired by the current router.
