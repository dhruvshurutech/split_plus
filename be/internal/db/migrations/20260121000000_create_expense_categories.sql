-- +goose Up
-- +goose StatementBegin
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    icon TEXT,
    color TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL REFERENCES users(id),
    deleted_at TIMESTAMPTZ
);

-- Unique constraint: slugs must be unique per group
CREATE UNIQUE INDEX idx_expense_categories_group_slug 
    ON expense_categories(group_id, slug) 
    WHERE deleted_at IS NULL;

CREATE INDEX idx_expense_categories_group_id 
    ON expense_categories(group_id) 
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS expense_categories;
-- +goose StatementEnd
