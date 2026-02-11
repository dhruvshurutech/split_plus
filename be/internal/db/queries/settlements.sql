-- name: CreateSettlement :one
INSERT INTO settlements (
    group_id, type,
    payer_id, payer_pending_user_id,
    payee_id, payee_pending_user_id,
    amount, currency_code, status,
    payment_method, transaction_reference, notes, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
RETURNING *;

-- name: GetSettlementByID :one
SELECT * FROM settlements
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListSettlementsByGroup :many
SELECT
    s.id,
    s.group_id,
    s.payer_id,
    s.payer_pending_user_id,
    s.payee_id,
    s.payee_pending_user_id,
    s.amount,
    s.currency_code,
    s.status,
    s.payment_method,
    s.transaction_reference,
    s.paid_at,
    s.notes,
    s.created_at,
    s.created_by,
    s.updated_at,
    s.updated_by,
    COALESCE(payer_user.email, payer_pending.email) AS payer_email,
    COALESCE(payer_user.name, payer_pending.name) AS payer_name,
    payer_user.avatar_url AS payer_avatar_url,
    (s.payer_pending_user_id IS NOT NULL) AS payer_is_pending,
    COALESCE(payee_user.email, payee_pending.email) AS payee_email,
    COALESCE(payee_user.name, payee_pending.name) AS payee_name,
    payee_user.avatar_url AS payee_avatar_url,
    (s.payee_pending_user_id IS NOT NULL) AS payee_is_pending
FROM settlements s
LEFT JOIN users payer_user ON s.payer_id = payer_user.id
LEFT JOIN pending_users payer_pending ON s.payer_pending_user_id = payer_pending.id
LEFT JOIN users payee_user ON s.payee_id = payee_user.id
LEFT JOIN pending_users payee_pending ON s.payee_pending_user_id = payee_pending.id
WHERE s.group_id = $1 AND s.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: ListSettlementsByUser :many
SELECT
    s.id,
    s.group_id,
    s.payer_id,
    s.payer_pending_user_id,
    s.payee_id,
    s.payee_pending_user_id,
    s.amount,
    s.currency_code,
    s.status,
    s.payment_method,
    s.transaction_reference,
    s.paid_at,
    s.notes,
    s.created_at,
    s.created_by,
    s.updated_at,
    s.updated_by,
    g.name AS group_name,
    COALESCE(payer_user.email, payer_pending.email) AS payer_email,
    COALESCE(payer_user.name, payer_pending.name) AS payer_name,
    payer_user.avatar_url AS payer_avatar_url,
    (s.payer_pending_user_id IS NOT NULL) AS payer_is_pending,
    COALESCE(payee_user.email, payee_pending.email) AS payee_email,
    COALESCE(payee_user.name, payee_pending.name) AS payee_name,
    payee_user.avatar_url AS payee_avatar_url,
    (s.payee_pending_user_id IS NOT NULL) AS payee_is_pending
FROM settlements s
JOIN groups g ON s.group_id = g.id
LEFT JOIN users payer_user ON s.payer_id = payer_user.id
LEFT JOIN pending_users payer_pending ON s.payer_pending_user_id = payer_pending.id
LEFT JOIN users payee_user ON s.payee_id = payee_user.id
LEFT JOIN pending_users payee_pending ON s.payee_pending_user_id = payee_pending.id
WHERE (
    s.payer_id = $1 OR
    s.payee_id = $1 OR
    s.payer_pending_user_id = $1 OR
    s.payee_pending_user_id = $1
) AND s.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: HasPendingMemberInvitation :one
SELECT EXISTS (
    SELECT 1
    FROM group_invitations gi
    JOIN pending_users pu ON pu.email = gi.email
    WHERE gi.group_id = $1
      AND pu.id = $2
      AND gi.status = 'pending'
      AND gi.expires_at > NOW()
);

-- name: UpdateSettlementStatus :one
UPDATE settlements
SET status = $2,
    paid_at = CASE WHEN $2 = 'completed' AND paid_at IS NULL THEN CURRENT_TIMESTAMP ELSE paid_at END,
    updated_at = CURRENT_TIMESTAMP,
    updated_by = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateSettlement :one
UPDATE settlements
SET amount = $2,
    currency_code = $3,
    status = $4,
    payment_method = $5,
    transaction_reference = $6,
    notes = $7,
    paid_at = CASE WHEN $4 = 'completed' AND paid_at IS NULL THEN CURRENT_TIMESTAMP ELSE paid_at END,
    updated_at = CURRENT_TIMESTAMP,
    updated_by = $8
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteSettlement :exec
UPDATE settlements
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListFriendSettlements :many
SELECT
    s.id,
    s.group_id,
    s.payer_id,
    s.payee_id,
    s.amount,
    s.currency_code,
    s.status,
    s.payment_method,
    s.transaction_reference,
    s.paid_at,
    s.notes,
    s.created_at,
    s.created_by,
    s.updated_at,
    s.updated_by,
    payer.email AS payer_email,
    payer.name AS payer_name,
    payer.avatar_url AS payer_avatar_url,
    payee.email AS payee_email,
    payee.name AS payee_name,
    payee.avatar_url AS payee_avatar_url
FROM settlements s
JOIN users payer ON s.payer_id = payer.id
JOIN users payee ON s.payee_id = payee.id
WHERE s.type = 'friend'
  AND s.group_id IS NULL
  AND ((s.payer_id = $1 AND s.payee_id = $2) OR (s.payer_id = $2 AND s.payee_id = $1))
  AND s.deleted_at IS NULL
ORDER BY s.created_at DESC;
