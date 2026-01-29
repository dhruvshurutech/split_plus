-- +goose Up
-- +goose StatementBegin
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id),
    title TEXT NOT NULL,
    notes TEXT,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    currency_code TEXT NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL REFERENCES users(id),
    deleted_at TIMESTAMPTZ
);

-- Index for querying expenses by group
CREATE INDEX idx_expenses_group_id ON expenses(group_id) WHERE deleted_at IS NULL;

-- Index for querying expenses by date
CREATE INDEX idx_expenses_date ON expenses(date) WHERE deleted_at IS NULL;

-- Index for querying expenses by creator
CREATE INDEX idx_expenses_created_by ON expenses(created_by) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS expenses;
-- +goose StatementEnd
