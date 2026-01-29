-- +goose Up
-- +goose StatementBegin
-- Add type column to settlements table
ALTER TABLE settlements
ADD COLUMN type TEXT NOT NULL DEFAULT 'group' CHECK (type IN ('group', 'friend'));

-- Make group_id nullable for friend settlements
ALTER TABLE settlements
ALTER COLUMN group_id DROP NOT NULL;

-- Add index for friend settlements
CREATE INDEX idx_settlements_type ON settlements(type) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_settlements_type;
ALTER TABLE settlements
ALTER COLUMN group_id SET NOT NULL;
ALTER TABLE settlements
DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
