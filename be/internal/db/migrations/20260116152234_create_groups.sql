-- +goose Up
-- +goose StatementBegin
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name TEXT NOT NULL,
    description TEXT,
    currency_code TEXT NOT NULL DEFAULT 'USD',
    default_split_method TEXT NOT NULL DEFAULT 'equal',
    settings JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL REFERENCES users(id),
    deleted_at TIMESTAMPTZ
);

-- Index for querying groups by currency_code
CREATE INDEX idx_groups_currency ON groups(currency_code) WHERE deleted_at IS NULL;

-- GIN index for settings JSONB field
CREATE INDEX idx_groups_settings ON groups USING GIN(settings);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS groups;
-- +goose StatementEnd
