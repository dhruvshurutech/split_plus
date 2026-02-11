-- name: GetUserThemePreferences :one
SELECT user_id, active_type, active_preset_slug, active_user_theme_id, mode, created_at, updated_at
FROM user_theme_preferences
WHERE user_id = $1;

-- name: UpsertUserThemePreferences :one
INSERT INTO user_theme_preferences (
    user_id,
    active_type,
    active_preset_slug,
    active_user_theme_id,
    mode
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id)
DO UPDATE SET
    active_type = EXCLUDED.active_type,
    active_preset_slug = EXCLUDED.active_preset_slug,
    active_user_theme_id = EXCLUDED.active_user_theme_id,
    mode = EXCLUDED.mode,
    updated_at = CURRENT_TIMESTAMP
RETURNING user_id, active_type, active_preset_slug, active_user_theme_id, mode, created_at, updated_at;
