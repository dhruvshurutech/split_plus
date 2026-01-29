-- name: CreateRecurringExpense :one
INSERT INTO recurring_expenses (
    group_id, title, notes, amount, currency_code, repeat_interval,
    day_of_month, day_of_week, start_date, end_date, next_occurrence_date,
    is_active, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
RETURNING *;

-- name: GetRecurringExpenseByID :one
SELECT * FROM recurring_expenses
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListRecurringExpensesByGroup :many
SELECT * FROM recurring_expenses
WHERE group_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateRecurringExpense :one
UPDATE recurring_expenses
SET title = $2,
    notes = $3,
    amount = $4,
    currency_code = $5,
    repeat_interval = $6,
    day_of_month = $7,
    day_of_week = $8,
    start_date = $9,
    end_date = $10,
    next_occurrence_date = $11,
    is_active = $12,
    updated_by = $13,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteRecurringExpense :exec
UPDATE recurring_expenses
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetRecurringExpensesDue :many
SELECT * FROM recurring_expenses
WHERE is_active = true
    AND deleted_at IS NULL
    AND next_occurrence_date <= CURRENT_DATE
ORDER BY next_occurrence_date ASC;

-- name: UpdateNextOccurrenceDate :one
UPDATE recurring_expenses
SET next_occurrence_date = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateRecurringExpenseActiveStatus :one
UPDATE recurring_expenses
SET is_active = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CreateRecurringExpensePayment :one
INSERT INTO recurring_expense_payments (recurring_expense_id, user_id, amount, payment_method)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListRecurringExpensePayments :many
SELECT
    rep.id,
    rep.recurring_expense_id,
    rep.user_id,
    rep.amount,
    rep.payment_method,
    rep.created_at,
    rep.updated_at,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url
FROM recurring_expense_payments rep
JOIN users u ON rep.user_id = u.id
WHERE rep.recurring_expense_id = $1 AND rep.deleted_at IS NULL
ORDER BY rep.created_at ASC;

-- name: CreateRecurringExpenseSplit :one
INSERT INTO recurring_expense_splits (recurring_expense_id, user_id, amount_owned, split_type)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListRecurringExpenseSplits :many
SELECT
    res.id,
    res.recurring_expense_id,
    res.user_id,
    res.amount_owned,
    res.split_type,
    res.created_at,
    res.updated_at,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url
FROM recurring_expense_splits res
JOIN users u ON res.user_id = u.id
WHERE res.recurring_expense_id = $1 AND res.deleted_at IS NULL
ORDER BY res.created_at ASC;
