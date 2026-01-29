-- +goose Up
-- +goose StatementBegin
CREATE TABLE settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id),
    payer_id UUID NOT NULL REFERENCES users(id),
    payee_id UUID NOT NULL REFERENCES users(id),
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    currency_code TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled')),
    payment_method TEXT,
    transaction_reference TEXT,
    paid_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    CHECK (payer_id != payee_id)
);

-- Indexes for settlements
CREATE INDEX idx_settlements_group_id ON settlements(group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_payer_id ON settlements(payer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_payee_id ON settlements(payee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_status ON settlements(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_paid_at ON settlements(paid_at DESC) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS settlements;
-- +goose StatementEnd
