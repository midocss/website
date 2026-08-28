-- Currencies are a first-class table from day one so multi-currency support can
-- be enabled later without touching every price column.
CREATE TABLE currencies (
    code          CHAR(3) PRIMARY KEY,
    name_ar       VARCHAR(64) NOT NULL,
    name_en       VARCHAR(64) NOT NULL,
    symbol        VARCHAR(8) NOT NULL,
    -- rate relative to the default currency
    exchange_rate NUMERIC(18, 6) NOT NULL DEFAULT 1,
    is_default    BOOLEAN NOT NULL DEFAULT FALSE,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_currencies_single_default ON currencies (is_default) WHERE is_default;

INSERT INTO currencies (code, name_ar, name_en, symbol, exchange_rate, is_default) VALUES
    ('IQD', 'دينار عراقي', 'Iraqi Dinar', 'د.ع', 1, TRUE),
    ('USD', 'دولار أمريكي', 'US Dollar', '$', 0.00076, FALSE);

CREATE TABLE project_types (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug           VARCHAR(96) NOT NULL UNIQUE,
    name_ar        VARCHAR(128) NOT NULL,
    name_en        VARCHAR(128) NOT NULL,
    description_ar TEXT,
    description_en TEXT,
    color_hex      CHAR(7) NOT NULL CHECK (color_hex ~* '^#[0-9a-f]{6}$'),
    icon_object_key TEXT,
    sort_order     INTEGER NOT NULL DEFAULT 0,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE TABLE portfolio_projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_type_id UUID REFERENCES project_types (id) ON DELETE SET NULL,
    slug            VARCHAR(160) NOT NULL UNIQUE,
    title_ar        VARCHAR(200) NOT NULL,
    title_en        VARCHAR(200) NOT NULL,
    description_ar  TEXT,
    description_en  TEXT,
    external_url    TEXT,
    cover_object_key TEXT,
    completed_at    DATE,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    is_published    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_portfolio_projects_type ON portfolio_projects (project_type_id);
CREATE INDEX idx_portfolio_projects_published ON portfolio_projects (is_published) WHERE deleted_at IS NULL;

CREATE TABLE portfolio_project_images (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_project_id UUID NOT NULL REFERENCES portfolio_projects (id) ON DELETE CASCADE,
    object_key           TEXT NOT NULL,
    alt_ar               VARCHAR(200),
    alt_en               VARCHAR(200),
    sort_order           INTEGER NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_portfolio_images_project ON portfolio_project_images (portfolio_project_id);

CREATE TABLE packages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_type_id UUID REFERENCES project_types (id) ON DELETE SET NULL,
    slug            VARCHAR(160) NOT NULL UNIQUE,
    name_ar         VARCHAR(160) NOT NULL,
    name_en         VARCHAR(160) NOT NULL,
    description_ar  TEXT,
    description_en  TEXT,
    price_amount    NUMERIC(14, 2) NOT NULL CHECK (price_amount >= 0),
    currency_code   CHAR(3) NOT NULL REFERENCES currencies (code) ON DELETE RESTRICT,
    delivery_days   INTEGER CHECK (delivery_days IS NULL OR delivery_days > 0),
    is_featured     BOOLEAN NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_packages_type ON packages (project_type_id);

CREATE TABLE package_features (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
    text_ar    VARCHAR(255) NOT NULL,
    text_en    VARCHAR(255) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_package_features_package ON package_features (package_id);
