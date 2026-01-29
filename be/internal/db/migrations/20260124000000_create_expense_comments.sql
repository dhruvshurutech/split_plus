-- +goose Up
CREATE TABLE expense_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL CHECK (length(comment) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_expense_comments_expense_id ON expense_comments(expense_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_comments_user_id ON expense_comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_comments_created_at ON expense_comments(created_at DESC) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS expense_comments;
