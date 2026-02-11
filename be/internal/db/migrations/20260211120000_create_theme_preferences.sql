-- +goose Up
-- +goose StatementBegin
CREATE TABLE theme_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    slug TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    family TEXT NOT NULL CHECK (family IN ('palette', 'expressive')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_themes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID NOT NULL REFERENCES users (id),
    name TEXT NOT NULL,
    base_preset_slug TEXT NOT NULL REFERENCES theme_presets (slug),
    font_family_key TEXT NOT NULL,
    light_tokens JSONB NOT NULL DEFAULT '{}'::jsonb,
    dark_tokens JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_user_themes_user_id_name_active
    ON user_themes (user_id, lower(name))
    WHERE deleted_at IS NULL;
CREATE INDEX idx_user_themes_user_id_active ON user_themes (user_id) WHERE deleted_at IS NULL;

CREATE TABLE user_theme_preferences (
    user_id UUID PRIMARY KEY REFERENCES users (id),
    active_type TEXT NOT NULL CHECK (active_type IN ('preset', 'custom')),
    active_preset_slug TEXT REFERENCES theme_presets (slug),
    active_user_theme_id UUID REFERENCES user_themes (id),
    mode TEXT NOT NULL CHECK (mode IN ('light', 'dark')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        (active_type = 'preset' AND active_preset_slug IS NOT NULL AND active_user_theme_id IS NULL) OR
        (active_type = 'custom' AND active_preset_slug IS NULL AND active_user_theme_id IS NOT NULL)
    )
);

INSERT INTO theme_presets (slug, label, family) VALUES
('minimal', 'Minimal', 'palette'),
('soft-professional', 'Soft Pro', 'palette'),
('atelier', 'Atelier', 'palette'),
('newspaper', 'Newspaper', 'expressive'),
('mono-slate', 'Mono Slate', 'expressive');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_theme_preferences;
DROP TABLE IF EXISTS user_themes;
DROP TABLE IF EXISTS theme_presets;
-- +goose StatementEnd
