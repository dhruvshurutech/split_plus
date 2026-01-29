-- +goose Up
CREATE TABLE pending_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE group_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL DEFAULT 'member',
    status TEXT NOT NULL DEFAULT 'pending',
    invited_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE expense_payments ADD COLUMN pending_user_id UUID REFERENCES pending_users(id);
ALTER TABLE expense_split ADD COLUMN pending_user_id UUID REFERENCES pending_users(id);

-- Constraint: Payment must have either user_id OR pending_user_id, but not both (or both? no, usually one).
-- Actually, strict XOR might be hard to enforce if we allow "claiming" later.
-- But for now, let's just allow nullable user_id.
-- Currently expense_payments.user_id is NOT NULL?
-- I need to check `expenses.sql` or schema.
-- If user_id is NOT NULL, I must make it nullable.

ALTER TABLE expense_payments ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE expense_split ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE expense_payments ADD CONSTRAINT payment_user_xor_check CHECK (
    (user_id IS NOT NULL AND pending_user_id IS NULL) OR
    (user_id IS NULL AND pending_user_id IS NOT NULL)
);

ALTER TABLE expense_split ADD CONSTRAINT split_user_xor_check CHECK (
    (user_id IS NOT NULL AND pending_user_id IS NULL) OR
    (user_id IS NULL AND pending_user_id IS NOT NULL)
);

-- +goose Down
ALTER TABLE expense_split DROP CONSTRAINT split_user_xor_check;
ALTER TABLE expense_payments DROP CONSTRAINT payment_user_xor_check;

-- Warning: Data loss if we drop nullable and there are NULLs.
-- We assume rollback happens only in dev/testing or we verify data.
-- Updating user_id to NOT NULL requires cleaning up pending users?
-- For simple down migration, we just drop columns and tables.
-- But we can't easily revert NULLable user_id if there are rows with NULL user_id.
-- We will just try to revert structure.

ALTER TABLE expense_split DROP COLUMN pending_user_id;
ALTER TABLE expense_payments DROP COLUMN pending_user_id;

DROP TABLE group_invitations;
DROP TABLE pending_users;
