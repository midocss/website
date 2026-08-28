# midocss platform — backend

Go monolith backing the public website, the portfolio, the quotes system, the
digital store and the admin dashboard.

- Requirements and roadmap: [`backend-prompt-en.md`](backend-prompt-en.md)
- Architecture and design decisions: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- Database schema: [`docs/SCHEMA.md`](docs/SCHEMA.md)
- API reference: [`docs/openapi.yaml`](docs/openapi.yaml)

**Stack:** Go 1.27 · Gin · GORM · PostgreSQL 16 · golang-migrate · JWT · MinIO (planned) · ZainCash (planned)

## Status

| Phase | Scope | State |
| --- | --- | --- |
| 1 | Project skeleton, config, DB, router, middlewares, full schema migrations, auth + permissions | done |
| 2 | Project types, portfolio, packages (CRUD + MinIO SVG upload) | next |
| 3 | Quotes system + notification outbox (email/WhatsApp) | planned |
| 4 | Store: digital products, cart, orders | planned |
| 5 | Payments (ZainCash) + PDF invoices + signed downloads | planned |

## Getting started

```bash
cp .env.example .env
# JWT_SECRET must be at least 32 characters
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$(openssl rand -hex 32)|" .env

make infra-up                      # PostgreSQL + MinIO via docker compose
make migrate-up                    # requires the golang-migrate CLI
SEED_ADMIN_EMAIL=admin@example.com SEED_ADMIN_PASSWORD='ChangeMe123' make seed
make run                           # http://localhost:8080
```

Installing the migration CLI:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Common commands

```bash
make test      # go test ./... -race -cover
make vet
make fmt
make build     # bin/api and bin/seed
make migrate-create name=add_something
```

## Quick check

```bash
curl -s localhost:8080/health/ready

curl -s -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Test User","email":"user@example.com","password":"passw0rd123"}'

curl -s localhost:8080/api/v1/auth/me -H "Authorization: Bearer <access_token>"
```

Every response uses the same envelope:

```json
{ "success": true, "data": { }, "request_id": "..." }
{ "success": false, "error": { "code": "validation_error", "message": "...", "fields": [] }, "request_id": "..." }
```
