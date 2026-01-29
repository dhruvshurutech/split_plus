-- +goose Up
-- +goose StatementBegin
-- Add type column to expenses table
ALTER TABLE expenses
ADD COLUMN type TEXT NOT NULL DEFAULT 'group' CHECK (type IN ('group', 'friend'));

-- Make group_id nullable for friend expenses
ALTER TABLE expenses
ALTER COLUMN group_id DROP NOT NULL;

-- Add index for friend expenses
CREATE INDEX idx_expenses_type ON expenses(type) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_expenses_type;
ALTER TABLE expenses
ALTER COLUMN group_id SET NOT NULL;
ALTER TABLE expenses
DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
