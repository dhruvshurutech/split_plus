-- name: ListThemePresets :many
SELECT id, slug, label, family, is_active, created_at, updated_at
FROM theme_presets
WHERE is_active = TRUE
ORDER BY slug;

-- name: GetThemePresetBySlug :one
SELECT id, slug, label, family, is_active, created_at, updated_at
FROM theme_presets
WHERE slug = $1;
