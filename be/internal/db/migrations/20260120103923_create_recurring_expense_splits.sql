-- +goose Up
-- +goose StatementBegin
CREATE TABLE recurring_expense_splits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recurring_expense_id UUID NOT NULL REFERENCES recurring_expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    amount_owned DECIMAL(10, 2) NOT NULL CHECK (amount_owned >= 0),
    split_type TEXT NOT NULL DEFAULT 'equal' CHECK (split_type IN ('equal', 'percentage', 'fixed', 'custom')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_recurring_expense_splits_unique
    ON recurring_expense_splits(recurring_expense_id, user_id)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recurring_expense_splits;
-- +goose StatementEnd
