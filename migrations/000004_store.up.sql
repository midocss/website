CREATE TABLE digital_products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            VARCHAR(160) NOT NULL UNIQUE,
    name_ar         VARCHAR(160) NOT NULL,
    name_en         VARCHAR(160) NOT NULL,
    description_ar  TEXT,
    description_en  TEXT,
    kind            VARCHAR(32) NOT NULL CHECK (kind IN ('template', 'icon_pack', 'tool', 'other')),
    price_amount    NUMERIC(14, 2) NOT NULL CHECK (price_amount >= 0),
    currency_code   CHAR(3) NOT NULL REFERENCES currencies (code) ON DELETE RESTRICT,
    -- object in the private bucket; only ever served through a signed URL
    file_object_key TEXT NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL CHECK (file_size_bytes > 0),
    preview_object_key TEXT,
    version         VARCHAR(32),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE carts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    currency_code CHAR(3) NOT NULL REFERENCES currencies (code) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_carts_user ON carts (user_id);

CREATE TABLE cart_items (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id            UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    item_type          VARCHAR(16) NOT NULL CHECK (item_type IN ('package', 'digital_product')),
    package_id         UUID REFERENCES packages (id) ON DELETE CASCADE,
    digital_product_id UUID REFERENCES digital_products (id) ON DELETE CASCADE,
    quantity           INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price         NUMERIC(14, 2) NOT NULL CHECK (unit_price >= 0),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT cart_items_reference_exactly_one CHECK (
        (item_type = 'package' AND package_id IS NOT NULL AND digital_product_id IS NULL)
        OR (item_type = 'digital_product' AND digital_product_id IS NOT NULL AND package_id IS NULL)
    )
);

CREATE UNIQUE INDEX idx_cart_items_package ON cart_items (cart_id, package_id) WHERE package_id IS NOT NULL;
CREATE UNIQUE INDEX idx_cart_items_product ON cart_items (cart_id, digital_product_id) WHERE digital_product_id IS NOT NULL;

CREATE TABLE orders (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    number         VARCHAR(32) NOT NULL UNIQUE,
    status         VARCHAR(24) NOT NULL DEFAULT 'pending_payment'
        CHECK (status IN ('pending_payment', 'paid', 'in_progress', 'completed', 'cancelled', 'refunded')),
    subtotal_amount NUMERIC(14, 2) NOT NULL CHECK (subtotal_amount >= 0),
    discount_amount NUMERIC(14, 2) NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    total_amount    NUMERIC(14, 2) NOT NULL CHECK (total_amount >= 0),
    currency_code   CHAR(3) NOT NULL REFERENCES currencies (code) ON DELETE RESTRICT,
    -- rate captured at checkout so historical totals stay reproducible
    exchange_rate   NUMERIC(18, 6) NOT NULL DEFAULT 1,
    customer_note   TEXT,
    placed_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user ON orders (user_id, placed_at DESC);
CREATE INDEX idx_orders_status ON orders (status);

CREATE TABLE order_items (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id           UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    item_type          VARCHAR(16) NOT NULL CHECK (item_type IN ('package', 'digital_product')),
    package_id         UUID REFERENCES packages (id) ON DELETE SET NULL,
    digital_product_id UUID REFERENCES digital_products (id) ON DELETE SET NULL,
    -- snapshot of the product at purchase time; the catalog may change later
    name_ar            VARCHAR(160) NOT NULL,
    name_en            VARCHAR(160) NOT NULL,
    unit_price         NUMERIC(14, 2) NOT NULL CHECK (unit_price >= 0),
    quantity           INTEGER NOT NULL CHECK (quantity > 0),
    total_price        NUMERIC(14, 2) NOT NULL CHECK (total_price >= 0),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_items_order ON order_items (order_id);
