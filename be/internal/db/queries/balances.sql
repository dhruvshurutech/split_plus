-- name: GetGroupBalances :many
-- Calculate balance for each user in a group
-- Balance = total_paid - total_owed
-- Positive balance means user is owed money, negative means user owes money
SELECT
    u.id AS user_id,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url,
    COALESCE(SUM(ep.amount), 0::DECIMAL) AS total_paid,
    COALESCE(SUM(es.amount_owned), 0::DECIMAL) AS total_owed
FROM group_members gm
JOIN users u ON gm.user_id = u.id
LEFT JOIN expenses e ON e.group_id = gm.group_id AND e.deleted_at IS NULL
LEFT JOIN expense_payments ep ON ep.expense_id = e.id AND ep.user_id = u.id AND ep.deleted_at IS NULL
LEFT JOIN expense_split es ON es.expense_id = e.id AND es.user_id = u.id AND es.deleted_at IS NULL
WHERE gm.group_id = $1 
    AND gm.status = 'active' 
    AND gm.deleted_at IS NULL
    AND u.deleted_at IS NULL
GROUP BY u.id, u.email, u.name, u.avatar_url
ORDER BY u.email;

-- name: GetUserBalanceInGroup :one
-- Get balance for a specific user in a group
SELECT
    u.id AS user_id,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url,
    COALESCE(SUM(ep.amount), 0::DECIMAL) AS total_paid,
    COALESCE(SUM(es.amount_owned), 0::DECIMAL) AS total_owed,
    COALESCE(SUM(ep.amount), 0::DECIMAL) - COALESCE(SUM(es.amount_owned), 0::DECIMAL) AS balance
FROM group_members gm
JOIN users u ON gm.user_id = u.id
LEFT JOIN expenses e ON e.group_id = gm.group_id AND e.deleted_at IS NULL
LEFT JOIN expense_payments ep ON ep.expense_id = e.id AND ep.user_id = u.id AND ep.deleted_at IS NULL
LEFT JOIN expense_split es ON es.expense_id = e.id AND es.user_id = u.id AND es.deleted_at IS NULL
WHERE gm.group_id = $1 
    AND gm.user_id = $2
    AND gm.status = 'active' 
    AND gm.deleted_at IS NULL
    AND u.deleted_at IS NULL
GROUP BY u.id, u.email, u.name, u.avatar_url;

-- name: GetOverallUserBalance :many
-- Get user's balance across all groups
SELECT
    g.id AS group_id,
    g.name AS group_name,
    g.currency_code AS currency_code,
    COALESCE(SUM(ep.amount), 0::DECIMAL) AS total_paid,
    COALESCE(SUM(es.amount_owned), 0::DECIMAL) AS total_owed
FROM group_members gm
JOIN groups g ON gm.group_id = g.id
LEFT JOIN expenses e ON e.group_id = gm.group_id AND e.deleted_at IS NULL
LEFT JOIN expense_payments ep ON ep.expense_id = e.id AND ep.user_id = gm.user_id AND ep.deleted_at IS NULL
LEFT JOIN expense_split es ON es.expense_id = e.id AND es.user_id = gm.user_id AND es.deleted_at IS NULL
WHERE gm.user_id = $1
    AND gm.status = 'active'
    AND gm.deleted_at IS NULL
    AND g.deleted_at IS NULL
GROUP BY g.id, g.name, g.currency_code
ORDER BY g.name;


