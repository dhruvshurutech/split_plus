-- name: ListCategoriesForGroup :many
SELECT * FROM expense_categories
WHERE group_id = $1 
  AND deleted_at IS NULL
ORDER BY name;

-- name: GetCategoryByID :one
SELECT * FROM expense_categories
WHERE id = $1 
  AND deleted_at IS NULL;

-- name: GetCategoryBySlug :one
SELECT * FROM expense_categories
WHERE group_id = $1 
  AND slug = $2 
  AND deleted_at IS NULL;

-- name: CreateGroupCategory :one
INSERT INTO expense_categories (
    group_id,
    slug,
    name,
    icon,
    color,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $6
) RETURNING *;

-- name: UpdateGroupCategory :one
UPDATE expense_categories
SET 
    slug = $2,
    name = $3,
    icon = $4,
    color = $5,
    updated_by = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteGroupCategory :exec
UPDATE expense_categories
SET 
    deleted_at = CURRENT_TIMESTAMP,
    updated_by = $2
WHERE id = $1 
  AND deleted_at IS NULL;
