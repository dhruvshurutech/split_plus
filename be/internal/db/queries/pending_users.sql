-- name: CreatePendingUser :one
INSERT INTO pending_users (email, name)
VALUES ($1, $2)
ON CONFLICT (email) DO UPDATE 
SET name = COALESCE(EXCLUDED.name, pending_users.name), 
    updated_at = NOW()
RETURNING *;

-- name: GetPendingUserByEmail :one
SELECT * FROM pending_users WHERE email = $1;

-- name: GetPendingUserByID :one
SELECT * FROM pending_users WHERE id = $1;

-- name: UpdatePendingPaymentUserID :exec
UPDATE expense_payments
SET user_id = $1, pending_user_id = NULL
WHERE pending_user_id = $2;

-- name: UpdatePendingSplitUserID :exec
UPDATE expense_split
SET user_id = $1, pending_user_id = NULL
WHERE pending_user_id = $2;

