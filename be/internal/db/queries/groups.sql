-- name: CreateGroup :one
INSERT INTO groups (name, description, currency_code, created_by, updated_by)
VALUES ($1, $2, $3, $4, $4) RETURNING *;

-- name: GetGroupByID :one
SELECT * FROM groups
WHERE id = $1 AND deleted_at IS NULL;