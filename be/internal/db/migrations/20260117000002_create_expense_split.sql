-- +goose Up
-- +goose StatementBegin
CREATE TABLE expense_split (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    amount_owned DECIMAL(10, 2) NOT NULL CHECK (amount_owned >= 0),
    split_type TEXT NOT NULL DEFAULT 'equal' CHECK (split_type IN ('equal', 'percentage', 'fixed', 'custom')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Unique constraint: a user can only have one split per expense (excluding soft-deleted records)
CREATE UNIQUE INDEX idx_expense_split_unique_split
    ON expense_split(expense_id, user_id)
    WHERE deleted_at IS NULL;

-- Index for querying splits by expense
CREATE INDEX idx_expense_split_expense_id ON expense_split(expense_id) WHERE deleted_at IS NULL;

-- Index for querying splits by user
CREATE INDEX idx_expense_split_user_id ON expense_split(user_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS expense_split;
-- +goose StatementEnd
