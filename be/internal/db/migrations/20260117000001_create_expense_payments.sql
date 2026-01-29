-- +goose Up
-- +goose StatementBegin
CREATE TABLE expense_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    payment_method TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Unique constraint: a user can only have one payment per expense (excluding soft-deleted records)
CREATE UNIQUE INDEX idx_expense_payments_unique_payment
    ON expense_payments(expense_id, user_id)
    WHERE deleted_at IS NULL;

-- Index for querying payments by expense
CREATE INDEX idx_expense_payments_expense_id ON expense_payments(expense_id) WHERE deleted_at IS NULL;

-- Index for querying payments by user
CREATE INDEX idx_expense_payments_user_id ON expense_payments(user_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS expense_payments;
-- +goose StatementEnd
