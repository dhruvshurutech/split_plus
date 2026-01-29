-- +goose Up
-- Full-text search on expense title and notes
CREATE INDEX idx_expenses_title_search ON expenses USING GIN(to_tsvector('english', title));
CREATE INDEX idx_expenses_notes_search ON expenses USING GIN(to_tsvector('english', notes)) WHERE notes IS NOT NULL;

-- Composite indexes for common filter combinations
CREATE INDEX idx_expenses_group_date_category ON expenses(group_id, date DESC, category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_amount_range ON expenses(group_id, amount) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_expenses_title_search;
DROP INDEX IF EXISTS idx_expenses_notes_search;
DROP INDEX IF EXISTS idx_expenses_group_date_category;
DROP INDEX IF EXISTS idx_expenses_amount_range;
