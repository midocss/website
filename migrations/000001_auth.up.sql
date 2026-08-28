CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(64) NOT NULL UNIQUE,
    name_ar     VARCHAR(128) NOT NULL,
    name_en     VARCHAR(128) NOT NULL,
    description TEXT,
    -- system roles (super_admin, customer) cannot be deleted from the dashboard
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(128) NOT NULL UNIQUE,
    resource    VARCHAR(64) NOT NULL,
    action      VARCHAR(64) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_permissions_resource ON permissions (resource);

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id       UUID NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
    email         VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    full_name     VARCHAR(160) NOT NULL,
    phone         VARCHAR(32),
    locale        VARCHAR(8) NOT NULL DEFAULT 'ar',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_email ON users (lower(email)) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role_id ON users (role_id);

-- Per-user overrides on top of the role permissions. `effect` allows granting an
-- extra permission to a single staff member or revoking one their role has.
CREATE TABLE user_permissions (
    user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    effect        VARCHAR(8) NOT NULL CHECK (effect IN ('allow', 'deny')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, permission_id)
);

CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    user_agent VARCHAR(255),
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    -- set when the token is rotated so a replayed token can invalidate the chain
    replaced_by UUID REFERENCES refresh_tokens (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

INSERT INTO roles (slug, name_ar, name_en, description, is_system) VALUES
    ('super_admin', 'مدير عام', 'Super Admin', 'Full unrestricted access', TRUE),
    ('staff', 'موظف', 'Staff', 'Dashboard user with granular permissions', TRUE),
    ('customer', 'عميل', 'Customer', 'Store customer account', TRUE);

INSERT INTO permissions (slug, resource, action, description) VALUES
    ('users.view', 'users', 'view', 'List and view users'),
    ('users.create', 'users', 'create', 'Create users'),
    ('users.update', 'users', 'update', 'Update users'),
    ('users.delete', 'users', 'delete', 'Delete users'),
    ('roles.view', 'roles', 'view', 'List and view roles'),
    ('roles.manage', 'roles', 'manage', 'Create, update and delete roles and their permissions'),
    ('project_types.view', 'project_types', 'view', 'View project types'),
    ('project_types.create', 'project_types', 'create', 'Create project types'),
    ('project_types.update', 'project_types', 'update', 'Update project types'),
    ('project_types.delete', 'project_types', 'delete', 'Delete project types'),
    ('portfolio.view', 'portfolio', 'view', 'View portfolio projects'),
    ('portfolio.create', 'portfolio', 'create', 'Create portfolio projects'),
    ('portfolio.update', 'portfolio', 'update', 'Update portfolio projects'),
    ('portfolio.delete', 'portfolio', 'delete', 'Delete portfolio projects'),
    ('packages.view', 'packages', 'view', 'View packages'),
    ('packages.create', 'packages', 'create', 'Create packages'),
    ('packages.update', 'packages', 'update', 'Update packages'),
    ('packages.delete', 'packages', 'delete', 'Delete packages'),
    ('quotes.view', 'quotes', 'view', 'View custom quote requests'),
    ('quotes.update', 'quotes', 'update', 'Update quote request status'),
    ('quotes.delete', 'quotes', 'delete', 'Delete quote requests'),
    ('products.view', 'products', 'view', 'View digital products'),
    ('products.create', 'products', 'create', 'Create digital products'),
    ('products.update', 'products', 'update', 'Update digital products'),
    ('products.delete', 'products', 'delete', 'Delete digital products'),
    ('orders.view', 'orders', 'view', 'View orders'),
    ('orders.update', 'orders', 'update', 'Update order status'),
    ('payments.view', 'payments', 'view', 'View payment transactions'),
    ('payments.refund', 'payments', 'refund', 'Refund payment transactions'),
    ('invoices.view', 'invoices', 'view', 'View and download invoices');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p WHERE r.slug = 'super_admin';
