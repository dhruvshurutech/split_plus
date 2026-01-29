-- name: CreateExpenseComment :one
INSERT INTO expense_comments (
    expense_id,
    user_id,
    comment
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetExpenseCommentByID :one
SELECT 
    c.*,
    u.name as user_name,
    u.email as user_email,
    u.avatar_url as user_avatar_url
FROM expense_comments c
JOIN users u ON c.user_id = u.id
WHERE c.id = $1 AND c.deleted_at IS NULL;

-- name: ListExpenseComments :many
SELECT 
    c.*,
    u.name as user_name,
    u.email as user_email,
    u.avatar_url as user_avatar_url
FROM expense_comments c
JOIN users u ON c.user_id = u.id
WHERE c.expense_id = $1 AND c.deleted_at IS NULL
ORDER BY c.created_at ASC;

-- name: UpdateExpenseComment :one
UPDATE expense_comments
SET 
    comment = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteExpenseComment :exec
UPDATE expense_comments
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CountExpenseComments :one
SELECT COUNT(*) FROM expense_comments
WHERE expense_id = $1 AND deleted_at IS NULL;
