-- name: ListUserThemes :many
SELECT id, user_id, name, base_preset_slug, font_family_key, light_tokens, dark_tokens, created_at, updated_at, deleted_at
FROM user_themes
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: GetUserThemeByID :one
SELECT id, user_id, name, base_preset_slug, font_family_key, light_tokens, dark_tokens, created_at, updated_at, deleted_at
FROM user_themes
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateUserTheme :one
INSERT INTO user_themes (
    user_id,
    name,
    base_preset_slug,
    font_family_key,
    light_tokens,
    dark_tokens
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, name, base_preset_slug, font_family_key, light_tokens, dark_tokens, created_at, updated_at, deleted_at;

-- name: UpdateUserTheme :one
UPDATE user_themes
SET
    name = $2,
    base_preset_slug = $3,
    font_family_key = $4,
    light_tokens = $5,
    dark_tokens = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, user_id, name, base_preset_slug, font_family_key, light_tokens, dark_tokens, created_at, updated_at, deleted_at;

-- name: DeleteUserTheme :exec
UPDATE user_themes
SET
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
