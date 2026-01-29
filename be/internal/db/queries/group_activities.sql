-- name: CreateGroupActivity :one
INSERT INTO group_activities (
    group_id,
    user_id,
    action,
    entity_type,
    entity_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListGroupActivities :many
SELECT 
    ga.id,
    ga.group_id,
    ga.user_id,
    ga.action,
    ga.entity_type,
    ga.entity_id,
    ga.metadata,
    ga.created_at,
    u.email as user_email,
    u.name as user_name,
    u.avatar_url as user_avatar_url
FROM group_activities ga
JOIN users u ON ga.user_id = u.id
WHERE ga.group_id = $1
ORDER BY ga.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetExpenseHistory :many
SELECT 
    ga.id,
    ga.group_id,
    ga.user_id,
    ga.action,
    ga.entity_type,
    ga.entity_id,
    ga.metadata,
    ga.created_at,
    u.email as user_email,
    u.name as user_name,
    u.avatar_url as user_avatar_url
FROM group_activities ga
JOIN users u ON ga.user_id = u.id
WHERE ga.entity_type = 'expense' 
  AND ga.entity_id = $1
ORDER BY ga.created_at DESC;
