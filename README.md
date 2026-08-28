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
| 2 | Project types, portfolio, packages: admin CRUD + public catalog endpoints | done (MinIO upload pending) |
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

# Public catalog (published/active rows only, no token needed)
curl -s localhost:8080/api/v1/project-types
curl -s 'localhost:8080/api/v1/portfolio?project_type=e-commerce-store&page=1&per_page=20'
curl -s localhost:8080/api/v1/portfolio/company-website
curl -s 'localhost:8080/api/v1/packages?featured=true'

# Dashboard CRUD (requires the matching permission)
curl -s -X POST localhost:8080/api/v1/admin/project-types \
  -H "Authorization: Bearer <access_token>" -H 'Content-Type: application/json' \
  -d '{"name_ar":"متجر إلكتروني","name_en":"E-Commerce Store","color_hex":"#3366FF"}'

curl -s -X POST localhost:8080/api/v1/admin/packages \
  -H "Authorization: Bearer <access_token>" -H 'Content-Type: application/json' \
  -d '{"name_ar":"الباقة الاحترافية","name_en":"Pro Package","price_amount":"1250000","currency_code":"IQD","features":[{"text_ar":"تصميم","text_en":"Design"}]}'
```

Slugs are generated from the English name when omitted, prices are decimal
strings validated against the active currencies, and portfolio projects start
unpublished so they stay out of the public endpoints until reviewed.

Every response uses the same envelope:

```json
{ "success": true, "data": { }, "request_id": "..." }
{ "success": false, "error": { "code": "validation_error", "message": "...", "fields": [] }, "request_id": "..." }
```
