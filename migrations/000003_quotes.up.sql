CREATE TABLE quote_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_type_id UUID REFERENCES project_types (id) ON DELETE SET NULL,
    -- set when the visitor was logged in as a customer
    user_id         UUID REFERENCES users (id) ON DELETE SET NULL,
    full_name       VARCHAR(160) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    whatsapp_number VARCHAR(32) NOT NULL,
    company         VARCHAR(160),
    budget_amount   NUMERIC(14, 2) CHECK (budget_amount IS NULL OR budget_amount >= 0),
    budget_currency CHAR(3) REFERENCES currencies (code) ON DELETE SET NULL,
    details         TEXT NOT NULL,
    preferred_lang  VARCHAR(8) NOT NULL DEFAULT 'ar',
    status          VARCHAR(24) NOT NULL DEFAULT 'new'
        CHECK (status IN ('new', 'under_review', 'responded', 'rejected')),
    admin_notes     TEXT,
    handled_by      UUID REFERENCES users (id) ON DELETE SET NULL,
    handled_at      TIMESTAMPTZ,
    source_ip       INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quote_requests_status ON quote_requests (status, created_at DESC);
CREATE INDEX idx_quote_requests_email ON quote_requests (lower(email));

CREATE TABLE quote_request_attachments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_request_id UUID NOT NULL REFERENCES quote_requests (id) ON DELETE CASCADE,
    object_key       TEXT NOT NULL,
    file_name        VARCHAR(255) NOT NULL,
    content_type     VARCHAR(128) NOT NULL,
    size_bytes       BIGINT NOT NULL CHECK (size_bytes > 0),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quote_attachments_request ON quote_request_attachments (quote_request_id);

-- Outbox for every outgoing message. The channel column keeps email and
-- WhatsApp (and any future channel) on the same delivery/retry pipeline.
CREATE TABLE notifications (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel       VARCHAR(24) NOT NULL CHECK (channel IN ('email', 'whatsapp', 'sms')),
    template      VARCHAR(96) NOT NULL,
    recipient     VARCHAR(255) NOT NULL,
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    status        VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'sent', 'failed')),
    attempts      INTEGER NOT NULL DEFAULT 0,
    last_error    TEXT,
    scheduled_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_pending ON notifications (status, scheduled_at);
