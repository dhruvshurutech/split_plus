-- name: CreateFriendship :one
INSERT INTO friendships (user_id, friend_user_id, status)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetFriendship :one
SELECT * FROM friendships
WHERE user_id = $1 AND friend_user_id = $2 AND deleted_at IS NULL;

-- name: GetFriendshipByID :one
SELECT * FROM friendships
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListFriends :many
SELECT
    f.id,
    f.user_id,
    f.friend_user_id,
    f.status,
    f.created_at,
    f.updated_at,
    u.id AS friend_id,
    u.email AS friend_email,
    u.name AS friend_name,
    u.avatar_url AS friend_avatar_url
FROM friendships f
JOIN users u
  ON (CASE WHEN f.user_id = $1 THEN f.friend_user_id ELSE f.user_id END) = u.id
WHERE (f.user_id = $1 OR f.friend_user_id = $1)
  AND f.status = 'accepted'
  AND f.deleted_at IS NULL
ORDER BY u.email;

-- name: ListIncomingFriendRequests :many
SELECT * FROM friendships
WHERE friend_user_id = $1
  AND status = 'pending'
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListOutgoingFriendRequests :many
SELECT * FROM friendships
WHERE user_id = $1
  AND status = 'pending'
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateFriendshipStatus :one
UPDATE friendships
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteFriendship :exec
UPDATE friendships
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL;

