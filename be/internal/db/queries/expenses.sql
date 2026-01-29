-- name: CreateExpense :one
INSERT INTO expenses (group_id, type, title, notes, amount, currency_code, date, category_id, tags, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10) RETURNING *;

-- name: GetExpenseByID :one
SELECT * FROM expenses
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListExpensesByGroup :many
SELECT * FROM expenses
WHERE group_id = $1 AND deleted_at IS NULL
ORDER BY date DESC, created_at DESC;

-- name: UpdateExpense :one
UPDATE expenses
SET title = $2,
    notes = $3,
    amount = $4,
    currency_code = $5,
    date = $6,
    category_id = $7,
    tags = $8,
    updated_by = $9,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteExpense :exec
UPDATE expenses
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateExpensePayment :one
INSERT INTO expense_payments (expense_id, user_id, pending_user_id, amount, payment_method)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: ListExpensePayments :many
SELECT
    ep.id,
    ep.expense_id,
    ep.user_id,
    ep.pending_user_id,
    ep.amount,
    ep.payment_method,
    ep.created_at,
    ep.updated_at,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url,
    pu.email AS pending_user_email,
    pu.name AS pending_user_name
FROM expense_payments ep
LEFT JOIN users u ON ep.user_id = u.id
LEFT JOIN pending_users pu ON ep.pending_user_id = pu.id
WHERE ep.expense_id = $1 AND ep.deleted_at IS NULL
ORDER BY ep.created_at ASC;

-- name: DeleteExpensePayments :exec
UPDATE expense_payments
SET deleted_at = CURRENT_TIMESTAMP
WHERE expense_id = $1 AND deleted_at IS NULL;

-- name: CreateExpenseSplit :one
INSERT INTO expense_split (expense_id, user_id, pending_user_id, amount_owned, split_type)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: ListExpenseSplits :many
SELECT
    es.id,
    es.expense_id,
    es.user_id,
    es.pending_user_id,
    es.amount_owned,
    es.split_type,
    es.created_at,
    es.updated_at,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url,
    pu.email AS pending_user_email,
    pu.name AS pending_user_name
FROM expense_split es
LEFT JOIN users u ON es.user_id = u.id
LEFT JOIN pending_users pu ON es.pending_user_id = pu.id
WHERE es.expense_id = $1 AND es.deleted_at IS NULL
ORDER BY es.created_at ASC;

-- name: UpdateExpenseSplit :one
UPDATE expense_split
SET amount_owned = $3,
    split_type = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE expense_id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteExpenseSplits :exec
UPDATE expense_split
SET deleted_at = CURRENT_TIMESTAMP
WHERE expense_id = $1 AND deleted_at IS NULL;

-- name: ListFriendExpenses :many
SELECT e.*
FROM expenses e
WHERE e.type = 'friend'
  AND e.group_id IS NULL
  AND e.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM expense_payments ep
    WHERE ep.expense_id = e.id
      AND ep.user_id IN ($1, $2)
      AND ep.deleted_at IS NULL
  )
  AND EXISTS (
    SELECT 1 FROM expense_split es
    WHERE es.expense_id = e.id
      AND es.user_id IN ($1, $2)
      AND es.deleted_at IS NULL
  )
  AND NOT EXISTS (
    SELECT 1 FROM expense_payments ep
    WHERE ep.expense_id = e.id
      AND ep.user_id NOT IN ($1, $2)
      AND ep.deleted_at IS NULL
  )
  AND NOT EXISTS (
    SELECT 1 FROM expense_split es
    WHERE es.expense_id = e.id
      AND es.user_id NOT IN ($1, $2)
      AND es.deleted_at IS NULL
  )
ORDER BY e.date DESC, e.created_at DESC;

-- name: SearchExpenses :many
SELECT e.* FROM expenses e
WHERE e.group_id = sqlc.arg('group_id')
  AND e.deleted_at IS NULL
  AND (
    sqlc.narg('query')::text IS NULL OR 
    to_tsvector('english', e.title || ' ' || COALESCE(e.notes, '')) @@ plainto_tsquery('english', sqlc.narg('query'))
  )
  AND (sqlc.narg('start_date')::date IS NULL OR e.date >= sqlc.narg('start_date'))
  AND (sqlc.narg('end_date')::date IS NULL OR e.date <= sqlc.narg('end_date'))
  AND (sqlc.narg('category_id')::uuid IS NULL OR e.category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('created_by')::uuid IS NULL OR e.created_by = sqlc.narg('created_by'))
  AND (sqlc.narg('min_amount')::numeric IS NULL OR e.amount >= sqlc.narg('min_amount'))
  AND (sqlc.narg('max_amount')::numeric IS NULL OR e.amount <= sqlc.narg('max_amount'))
  AND (sqlc.narg('payer_id')::uuid IS NULL OR EXISTS (
    SELECT 1 FROM expense_payments ep 
    WHERE ep.expense_id = e.id AND ep.user_id = sqlc.narg('payer_id') AND ep.deleted_at IS NULL
  ))
  AND (sqlc.narg('ower_id')::uuid IS NULL OR EXISTS (
    SELECT 1 FROM expense_split es 
    WHERE es.expense_id = e.id AND es.user_id = sqlc.narg('ower_id') AND es.deleted_at IS NULL
  ))
ORDER BY e.date DESC, e.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
