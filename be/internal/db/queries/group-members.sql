-- name: CreateGroupMember :one
INSERT INTO group_members (group_id, user_id, role, status, invited_by, invited_at, joined_at)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetGroupMember :one
SELECT * FROM group_members
WHERE group_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: UpdateGroupMemberStatus :one
UPDATE group_members
SET status = $3, joined_at = $4, updated_at = CURRENT_TIMESTAMP
WHERE group_id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: ListGroupMembers :many
SELECT
    gm.id,
    gm.group_id,
    gm.user_id,
    gm.role,
    gm.status,
    gm.invited_by,
    gm.invited_at,
    gm.joined_at,
    gm.created_at,
    gm.updated_at,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url
FROM group_members gm
JOIN users u ON gm.user_id = u.id
WHERE gm.group_id = $1
    AND gm.deleted_at IS NULL
    AND gm.status IN ('active', 'pending')
ORDER BY gm.created_at ASC;

-- name: GetGroupsByUserID :many
SELECT
    g.id,
    g.name,
    g.description,
    g.currency_code,
    g.created_at,
    gm.id as membership_id,
    gm.role as member_role,
    gm.status as member_status,
    gm.joined_at as member_joined_at
FROM group_members AS gm
INNER JOIN groups AS g ON gm.group_id = g.id
WHERE gm.user_id = $1 AND gm.deleted_at IS NULL AND g.deleted_at IS NULL
ORDER BY g.created_at DESC;