CREATE TABLE payment_transactions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id       UUID NOT NULL REFERENCES orders (id) ON DELETE RESTRICT,
    gateway        VARCHAR(32) NOT NULL,
    -- identifier returned by the gateway (ZainCash transaction id, etc.)
    gateway_ref    VARCHAR(128),
    status         VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'success', 'failed', 'cancelled', 'refunded')),
    amount         NUMERIC(14, 2) NOT NULL CHECK (amount >= 0),
    currency_code  CHAR(3) NOT NULL REFERENCES currencies (code) ON DELETE RESTRICT,
    redirect_url   TEXT,
    failure_reason TEXT,
    request_payload  JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_transactions_order ON payment_transactions (order_id);
CREATE UNIQUE INDEX idx_payment_transactions_gateway_ref ON payment_transactions (gateway, gateway_ref)
    WHERE gateway_ref IS NOT NULL;

-- Raw webhook deliveries, stored before processing so a callback can be
-- replayed and so duplicate deliveries are idempotent.
CREATE TABLE payment_webhook_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway        VARCHAR(32) NOT NULL,
    external_id    VARCHAR(128),
    transaction_id UUID REFERENCES payment_transactions (id) ON DELETE SET NULL,
    payload        JSONB NOT NULL,
    signature_ok   BOOLEAN NOT NULL DEFAULT FALSE,
    processed_at   TIMESTAMPTZ,
    process_error  TEXT,
    received_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_payment_webhook_dedup ON payment_webhook_events (gateway, external_id)
    WHERE external_id IS NOT NULL;

CREATE TABLE invoices (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id       UUID NOT NULL UNIQUE REFERENCES orders (id) ON DELETE RESTRICT,
    number         VARCHAR(32) NOT NULL UNIQUE,
    pdf_object_key TEXT,
    total_amount   NUMERIC(14, 2) NOT NULL CHECK (total_amount >= 0),
    currency_code  CHAR(3) NOT NULL REFERENCES currencies (code) ON DELETE RESTRICT,
    issued_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Entitlement of a customer to download a purchased digital product.
CREATE TABLE download_grants (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_item_id      UUID NOT NULL REFERENCES order_items (id) ON DELETE CASCADE,
    user_id            UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    digital_product_id UUID NOT NULL REFERENCES digital_products (id) ON DELETE RESTRICT,
    max_downloads      INTEGER CHECK (max_downloads IS NULL OR max_downloads > 0),
    download_count     INTEGER NOT NULL DEFAULT 0,
    expires_at         TIMESTAMPTZ,
    last_downloaded_at TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_download_grants_user ON download_grants (user_id);
CREATE UNIQUE INDEX idx_download_grants_order_item ON download_grants (order_item_id);
