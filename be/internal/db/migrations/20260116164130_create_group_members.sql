-- +goose Up
-- +goose StatementBegin
CREATE TABLE group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'inactive')),
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Unique constraint: a user can only be a member of a group once (excluding soft-deleted records)
CREATE UNIQUE INDEX idx_group_members_unique_membership
    ON group_members(group_id, user_id)
    WHERE deleted_at IS NULL;

-- Index for querying members by group
CREATE INDEX idx_group_members_group_id ON group_members(group_id) WHERE deleted_at IS NULL;

-- Index for querying groups by user
CREATE INDEX idx_group_members_user_id ON group_members(user_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS group_members;
-- +goose StatementEnd
