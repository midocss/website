# Architecture

## Why a monolith

The platform is one Go binary serving the public website API, the store and the
admin dashboard API. A single deployable keeps transactions (cart → order →
payment → invoice) in one database and one process. The package layout below
keeps each domain isolated, so a domain can be extracted into its own service
later without rewriting the business logic.

## Why Gin (and not Fiber)

Both are fast enough for this workload; the tie-breakers:

- Gin is built on `net/http`, so the whole standard ecosystem works directly:
  `httptest` for handler tests, standard middleware, `http.Server` timeouts and
  graceful shutdown, OpenTelemetry instrumentation. Fiber is built on
  `fasthttp`, which has its own request/response types and is incompatible with
  `net/http` handlers and much of the ecosystem.
- Payment gateway SDKs, S3/MinIO SDKs and webhook helpers all speak
  `net/http`.
- Gin's binding + `validator/v10` integration gives structured validation
  errors, which map cleanly onto the unified error format.

## Layering

```
HTTP handler  →  service        →  repository   →  PostgreSQL
(transport)      (business)        (data access)
```

- **Handlers** only parse/validate input, call one service method and render
  the response. No SQL, no business rules.
- **Services** own the business rules and transactions. They depend on
  repository *interfaces*, which is what makes them unit-testable without a
  database (see `internal/auth/service_test.go`).
- **Repositories** own all persistence. Every error leaving a repository is
  already an `*apperr.Error`.
- **Domain** holds the entities shared by the layers.

## Folder structure

```
cmd/
  api/            # HTTP server entrypoint
  seed/           # idempotent super-admin seeding
internal/
  config/         # env-driven configuration + validation
  domain/         # entities (users, roles, permissions, refresh tokens, ...)
  auth/           # registration, login, JWT, refresh-token rotation
  rbac/           # effective-permission resolution + permission slugs
  users/          # admin user/role/permission management
  platform/
    database/     # PostgreSQL connection pool
    logger/       # structured slog setup
  transport/
    http/         # router + server wiring
      handler/    # gin handlers
      middleware/ # request id, logging, recovery, CORS, rate limit, auth
      httpx/      # binding + validation-error translation
      response/   # single JSON envelope
pkg/
  apperr/         # error type shared by every layer
migrations/       # golang-migrate SQL migrations
docs/             # architecture, schema, OpenAPI
```

Planned packages for the next phases (same shape): `catalog/` (project types,
portfolio, packages), `quotes/`, `store/` (products, cart, orders),
`payments/` (gateway interface + ZainCash), `invoices/`, `storage/` (MinIO),
`notifications/` (email + WhatsApp outbox).

## Authentication

- **Access token:** HS256 JWT, 15 minutes, carries only `uid` and `role`.
  Permissions are deliberately *not* in the token: they are resolved per
  request, so revoking a permission takes effect immediately.
- **Refresh token:** 32 random bytes, only the SHA-256 hash is stored. Every
  refresh rotates the token and links the old row to its replacement. Presenting
  an already-revoked token is treated as a leak and revokes all of that user's
  sessions.
- **Passwords:** bcrypt cost 12, with a minimum-strength policy shared by
  registration, admin-created accounts and (later) password reset.

## Permissions model

Role-based with per-user overrides:

```
effective = role_permissions(user.role) ∪ user_permissions(allow) \ user_permissions(deny)
```

Super admin bypasses the join and gets every permission. This satisfies "a staff
member may add offers but not delete them" without creating a role per person,
and new permissions are just new rows — no code change beyond a slug constant.

## Error handling

Every layer returns `*apperr.Error` (code + client-safe message + optional field
errors). The HTTP layer maps the code to a status and renders:

```json
{
  "success": false,
  "error": { "code": "validation_error", "message": "...", "fields": [{"field": "email", "message": "..."}] },
  "request_id": "..."
}
```

Internal errors log their cause with the request id and never leak it.

## Bilingual content (ar/en)

Two options were considered for dynamic content:

| Approach | Pros | Cons |
| --- | --- | --- |
| **Dual columns** (`name_ar`, `name_en`) — chosen | Simple queries, no joins, easy validation, both languages always loaded together, trivial sorting/filtering per language | Adding a third language is a migration touching every table |
| Separate `translations` table | Unlimited languages, no schema change per language | Every read needs a join or a second query; harder constraints; more complex admin CRUD |

Only two languages are planned and they are both always shown in the dashboard,
so dual columns win on simplicity. If a third language ever appears, the
migration path is mechanical: create `translations(entity_type, entity_id,
locale, field, value)`, backfill from the `_ar`/`_en` columns, then drop them.

Static UI strings stay in the frontend. The API selects the response language
from `Accept-Language`, falling back to `APP_DEFAULT_LANG`.

## Future scalability

- **Multi-currency:** prices already reference a `currencies` table, and orders
  snapshot the `exchange_rate` used at checkout, so historical totals stay
  reproducible when rates change.
- **Timezones:** every timestamp is `TIMESTAMPTZ` stored in UTC; rendering is a
  presentation concern.
- **New payment gateways:** `payment_transactions.gateway` plus a Go
  `PaymentGateway` interface (create payment → redirect URL, verify callback,
  refund). ZainCash is the first implementation; SuperQi/Visa plugs in without
  touching order or invoice logic.
- **Digital delivery:** files live in the private MinIO bucket and are only ever
  served through short-lived signed URLs issued against a `download_grants` row.
