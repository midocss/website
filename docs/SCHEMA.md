# Database schema

PostgreSQL, migrated with [golang-migrate](https://github.com/golang-migrate/migrate).
All ids are UUIDs (`gen_random_uuid()`), all timestamps are `TIMESTAMPTZ` in UTC,
all money is `NUMERIC(14,2)` paired with a `currency_code`.

## Relationships

```
roles ──< role_permissions >── permissions ──< user_permissions >── users
users ──< refresh_tokens
users ──< carts ──< cart_items >── packages / digital_products
users ──< orders ──< order_items >── packages / digital_products
orders ──< payment_transactions ──< payment_webhook_events
orders ──1 invoices
order_items ──1 download_grants >── digital_products
project_types ──< portfolio_projects ──< portfolio_project_images
project_types ──< packages ──< package_features
project_types ──< quote_requests ──< quote_request_attachments
currencies ──< packages / digital_products / orders / payment_transactions / invoices
```

## Tables

### 000001 — identity & access
| Table | Purpose |
| --- | --- |
| `roles` | `super_admin`, `staff`, `customer` seeded as system roles |
| `permissions` | `resource.action` slugs (e.g. `packages.delete`) |
| `role_permissions` | permissions granted by a role |
| `users` | accounts; soft-deleted; unique on `lower(email)` |
| `user_permissions` | per-user `allow`/`deny` override on top of the role |
| `refresh_tokens` | SHA-256 hash only, with rotation chain (`replaced_by`) |

### 000002 — catalog
| Table | Purpose |
| --- | --- |
| `currencies` | IQD (default) + USD, with exchange rate |
| `project_types` | name/description (ar+en), hex color (CHECK-validated), animated SVG object key |
| `portfolio_projects` / `portfolio_project_images` | past work, published flag, gallery |
| `packages` / `package_features` | fixed offers per project type and their feature list |

### 000003 — quotes & notifications
| Table | Purpose |
| --- | --- |
| `quote_requests` | custom quote form; status `new → under_review → responded/rejected` |
| `quote_request_attachments` | reference files on MinIO |
| `notifications` | outbox for email/WhatsApp/SMS with retry bookkeeping |

### 000004 — store
| Table | Purpose |
| --- | --- |
| `digital_products` | templates/icon packs/tools; private-bucket object key |
| `carts` / `cart_items` | one cart per customer; an item is a package **or** a digital product (CHECK-enforced) |
| `orders` / `order_items` | checkout snapshot: name and unit price copied at purchase time |

### 000005 — payments, invoices, delivery
| Table | Purpose |
| --- | --- |
| `payment_transactions` | one row per attempt: gateway, ref, status, request/response payloads |
| `payment_webhook_events` | raw callbacks, deduplicated on `(gateway, external_id)` for idempotent processing |
| `invoices` | one per paid order, PDF object key, sequential number |
| `download_grants` | entitlement backing every signed download URL (expiry + download count) |

## Notes

- Catalog rows are soft-deleted (`deleted_at`) because orders reference them
  historically; order items also store a name/price snapshot so a deleted
  product never breaks an old invoice.
- `orders.exchange_rate` freezes the rate used at checkout.
- Cart uniqueness indexes prevent the same package/product being added twice
  instead of incrementing quantity.
