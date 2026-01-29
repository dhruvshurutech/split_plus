-- +goose Up
-- +goose StatementBegin
CREATE TABLE recurring_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id),
    title TEXT NOT NULL,
    notes TEXT,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    currency_code TEXT NOT NULL,
    repeat_interval TEXT NOT NULL CHECK (repeat_interval IN ('daily', 'weekly', 'monthly', 'yearly')),
    day_of_month INTEGER CHECK (day_of_month >= 1 AND day_of_month <= 31), -- For monthly/yearly (NULL for daily/weekly)
    day_of_week INTEGER CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sunday, 6=Saturday, for weekly (NULL for daily/monthly/yearly)
    start_date DATE NOT NULL,
    end_date DATE,
    next_occurrence_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL REFERENCES users(id),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_recurring_expenses_group_id ON recurring_expenses(group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_recurring_expenses_next_occurrence ON recurring_expenses(next_occurrence_date) WHERE deleted_at IS NULL AND is_active = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recurring_expenses;
-- +goose StatementEnd
