-- +goose Up
-- +goose StatementBegin
ALTER TABLE expenses ADD COLUMN category_id UUID REFERENCES expense_categories(id);
ALTER TABLE expenses ADD COLUMN tags TEXT[] DEFAULT '{}';

CREATE INDEX idx_expenses_category_id ON expenses(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_tags ON expenses USING GIN(tags);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE expenses DROP COLUMN IF EXISTS category_id;
ALTER TABLE expenses DROP COLUMN IF EXISTS tags;
-- +goose StatementEnd
