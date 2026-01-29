-- name: CreateInvitation :one
INSERT INTO group_invitations (group_id, email, token, role, status, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetInvitationByToken :one
SELECT gi.*, g.name as group_name 
FROM group_invitations gi
JOIN groups g ON gi.group_id = g.id
WHERE gi.token = $1 AND gi.status = 'pending' AND gi.expires_at > NOW();

-- name: GetInvitationByID :one
SELECT * FROM group_invitations
WHERE id = $1;

-- name: UpdateInvitationStatus :one
UPDATE group_invitations
SET status = $2, updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: ListInvitationsByGroup :many
SELECT * FROM group_invitations
WHERE group_id = $1 ORDER BY created_at DESC;

-- name: GetPendingInvitationsByEmail :many
SELECT gi.*, g.name as group_name, u.name as inviter_name
FROM group_invitations gi
JOIN groups g ON gi.group_id = g.id
JOIN users u ON gi.invited_by = u.id
WHERE gi.email = $1 AND gi.status = 'pending' AND gi.expires_at > NOW();
