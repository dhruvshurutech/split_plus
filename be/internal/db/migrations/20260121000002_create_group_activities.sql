-- +goose Up
-- +goose StatementBegin
CREATE TABLE group_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_activities_group_id ON group_activities(group_id, created_at DESC);
CREATE INDEX idx_group_activities_user_id ON group_activities(user_id, created_at DESC);
CREATE INDEX idx_group_activities_entity ON group_activities(entity_type, entity_id);
CREATE INDEX idx_group_activities_metadata ON group_activities USING GIN(metadata);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS group_activities;
-- +goose StatementEnd
