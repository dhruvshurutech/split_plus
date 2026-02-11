-- name: GetGroupBalances :many
-- Calculate balance for each user in a group
-- Balance = total_paid - total_owed
-- Positive balance means user is owed money, negative means user owes money
WITH payments AS (
    SELECT
        ep.user_id,
        SUM(ep.amount) AS total_paid
    FROM expense_payments ep
    JOIN expenses e ON e.id = ep.expense_id
    WHERE e.group_id = $1
      AND e.deleted_at IS NULL
      AND ep.deleted_at IS NULL
    GROUP BY ep.user_id
),
settlement_payments AS (
    SELECT
        s.payer_id AS user_id,
        SUM(s.amount) AS total_paid
    FROM settlements s
    WHERE s.group_id = $1
      AND s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
      AND s.payer_id IS NOT NULL
    GROUP BY s.payer_id
),
splits AS (
    SELECT
        es.user_id,
        SUM(es.amount_owned) AS total_owed
    FROM expense_split es
    JOIN expenses e ON e.id = es.expense_id
    WHERE e.group_id = $1
      AND e.deleted_at IS NULL
      AND es.deleted_at IS NULL
    GROUP BY es.user_id
),
settlement_splits AS (
    SELECT
        s.payee_id AS user_id,
        SUM(s.amount) AS total_owed
    FROM settlements s
    WHERE s.group_id = $1
      AND s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
      AND s.payee_id IS NOT NULL
    GROUP BY s.payee_id
)
SELECT
    u.id AS user_id,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url,
    (COALESCE(p.total_paid, 0::DECIMAL) + COALESCE(sp.total_paid, 0::DECIMAL))::TEXT AS total_paid,
    (COALESCE(s.total_owed, 0::DECIMAL) + COALESCE(ss.total_owed, 0::DECIMAL))::TEXT AS total_owed
FROM group_members gm
JOIN users u ON gm.user_id = u.id
LEFT JOIN payments p ON p.user_id = u.id
LEFT JOIN settlement_payments sp ON sp.user_id = u.id
LEFT JOIN splits s ON s.user_id = u.id
LEFT JOIN settlement_splits ss ON ss.user_id = u.id
WHERE gm.group_id = $1 
    AND gm.status = 'active' 
    AND gm.deleted_at IS NULL
    AND u.deleted_at IS NULL
ORDER BY u.email;

-- name: GetGroupBalancesWithPending :many
-- Calculate balance for each user or pending user in a group
-- Balance = total_paid - total_owed
-- Positive balance means entity is owed money, negative means entity owes money
WITH payments AS (
    SELECT
        COALESCE(ep.user_id, ep.pending_user_id) AS entity_id,
        CASE WHEN ep.user_id IS NOT NULL THEN 'user' ELSE 'pending' END AS entity_type,
        SUM(ep.amount) AS total_paid
    FROM expense_payments ep
    JOIN expenses e ON e.id = ep.expense_id
    WHERE e.group_id = $1
      AND e.deleted_at IS NULL
      AND ep.deleted_at IS NULL
    GROUP BY COALESCE(ep.user_id, ep.pending_user_id),
             CASE WHEN ep.user_id IS NOT NULL THEN 'user' ELSE 'pending' END
),
settlement_payments AS (
    SELECT
        COALESCE(s.payer_id, s.payer_pending_user_id) AS entity_id,
        CASE WHEN s.payer_id IS NOT NULL THEN 'user' ELSE 'pending' END AS entity_type,
        SUM(s.amount) AS total_paid
    FROM settlements s
    WHERE s.group_id = $1
      AND s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
    GROUP BY COALESCE(s.payer_id, s.payer_pending_user_id),
             CASE WHEN s.payer_id IS NOT NULL THEN 'user' ELSE 'pending' END
),
splits AS (
    SELECT
        COALESCE(es.user_id, es.pending_user_id) AS entity_id,
        CASE WHEN es.user_id IS NOT NULL THEN 'user' ELSE 'pending' END AS entity_type,
        SUM(es.amount_owned) AS total_owed
    FROM expense_split es
    JOIN expenses e ON e.id = es.expense_id
    WHERE e.group_id = $1
      AND e.deleted_at IS NULL
      AND es.deleted_at IS NULL
    GROUP BY COALESCE(es.user_id, es.pending_user_id),
             CASE WHEN es.user_id IS NOT NULL THEN 'user' ELSE 'pending' END
),
settlement_splits AS (
    SELECT
        COALESCE(s.payee_id, s.payee_pending_user_id) AS entity_id,
        CASE WHEN s.payee_id IS NOT NULL THEN 'user' ELSE 'pending' END AS entity_type,
        SUM(s.amount) AS total_owed
    FROM settlements s
    WHERE s.group_id = $1
      AND s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
    GROUP BY COALESCE(s.payee_id, s.payee_pending_user_id),
             CASE WHEN s.payee_id IS NOT NULL THEN 'user' ELSE 'pending' END
),
entities AS (
    SELECT gm.user_id AS user_id, NULL::uuid AS pending_user_id, 'user' AS entity_type
    FROM group_members gm
    WHERE gm.group_id = $1
      AND gm.status = 'active'
      AND gm.deleted_at IS NULL
    UNION
    SELECT NULL::uuid AS user_id, pu.id AS pending_user_id, 'pending' AS entity_type
    FROM pending_users pu
    JOIN expense_split es ON es.pending_user_id = pu.id AND es.deleted_at IS NULL
    JOIN expenses e ON e.id = es.expense_id AND e.deleted_at IS NULL
    WHERE e.group_id = $1
    UNION
    SELECT NULL::uuid AS user_id, pu.id AS pending_user_id, 'pending' AS entity_type
    FROM pending_users pu
    JOIN expense_payments ep ON ep.pending_user_id = pu.id AND ep.deleted_at IS NULL
    JOIN expenses e ON e.id = ep.expense_id AND e.deleted_at IS NULL
    WHERE e.group_id = $1
    UNION
    SELECT NULL::uuid AS user_id, pu.id AS pending_user_id, 'pending' AS entity_type
    FROM pending_users pu
    JOIN settlements s ON s.payer_pending_user_id = pu.id
    WHERE s.group_id = $1
      AND s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
    UNION
    SELECT NULL::uuid AS user_id, pu.id AS pending_user_id, 'pending' AS entity_type
    FROM pending_users pu
    JOIN settlements s ON s.payee_pending_user_id = pu.id
    WHERE s.group_id = $1
      AND s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
)
SELECT
    ent.user_id AS user_id,
    ent.pending_user_id AS pending_user_id,
    COALESCE(u.email, pu.email) AS email,
    COALESCE(u.name, pu.name) AS name,
    u.avatar_url AS avatar_url,
    (COALESCE(p.total_paid, 0::DECIMAL) + COALESCE(sp.total_paid, 0::DECIMAL))::TEXT AS total_paid,
    (COALESCE(s.total_owed, 0::DECIMAL) + COALESCE(ss.total_owed, 0::DECIMAL))::TEXT AS total_owed
FROM entities ent
LEFT JOIN users u ON ent.user_id = u.id
LEFT JOIN pending_users pu ON ent.pending_user_id = pu.id
LEFT JOIN payments p
  ON p.entity_id = COALESCE(ent.user_id, ent.pending_user_id)
 AND p.entity_type = ent.entity_type
LEFT JOIN settlement_payments sp
  ON sp.entity_id = COALESCE(ent.user_id, ent.pending_user_id)
 AND sp.entity_type = ent.entity_type
LEFT JOIN splits s
  ON s.entity_id = COALESCE(ent.user_id, ent.pending_user_id)
 AND s.entity_type = ent.entity_type
LEFT JOIN settlement_splits ss
  ON ss.entity_id = COALESCE(ent.user_id, ent.pending_user_id)
 AND ss.entity_type = ent.entity_type
WHERE (u.id IS NULL OR u.deleted_at IS NULL)
ORDER BY COALESCE(u.email, pu.email);

-- name: GetUserBalanceInGroup :one
-- Get balance for a specific user in a group
WITH payments AS (
    SELECT
        ep.user_id,
        SUM(ep.amount) AS total_paid
    FROM expense_payments ep
    JOIN expenses e ON e.id = ep.expense_id
    WHERE e.group_id = $1
      AND e.deleted_at IS NULL
      AND ep.deleted_at IS NULL
    GROUP BY ep.user_id
),
settlement_payments AS (
    SELECT
        s.payer_id AS user_id,
        SUM(s.amount) AS total_paid
    FROM settlements s
    WHERE s.group_id = $1
      AND s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
      AND s.payer_id IS NOT NULL
    GROUP BY s.payer_id
),
splits AS (
    SELECT
        es.user_id,
        SUM(es.amount_owned) AS total_owed
    FROM expense_split es
    JOIN expenses e ON e.id = es.expense_id
    WHERE e.group_id = $1
      AND e.deleted_at IS NULL
      AND es.deleted_at IS NULL
    GROUP BY es.user_id
),
settlement_splits AS (
    SELECT
        s.payee_id AS user_id,
        SUM(s.amount) AS total_owed
    FROM settlements s
    WHERE s.group_id = $1
      AND s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
      AND s.payee_id IS NOT NULL
    GROUP BY s.payee_id
)
SELECT
    u.id AS user_id,
    u.email AS user_email,
    u.name AS user_name,
    u.avatar_url AS user_avatar_url,
    (COALESCE(p.total_paid, 0::DECIMAL) + COALESCE(sp.total_paid, 0::DECIMAL))::TEXT AS total_paid,
    (COALESCE(s.total_owed, 0::DECIMAL) + COALESCE(ss.total_owed, 0::DECIMAL))::TEXT AS total_owed,
    (
        (COALESCE(p.total_paid, 0::DECIMAL) + COALESCE(sp.total_paid, 0::DECIMAL)) -
        (COALESCE(s.total_owed, 0::DECIMAL) + COALESCE(ss.total_owed, 0::DECIMAL))
    )::TEXT AS balance
FROM group_members gm
JOIN users u ON gm.user_id = u.id
LEFT JOIN payments p ON p.user_id = u.id
LEFT JOIN settlement_payments sp ON sp.user_id = u.id
LEFT JOIN splits s ON s.user_id = u.id
LEFT JOIN settlement_splits ss ON ss.user_id = u.id
WHERE gm.group_id = $1 
    AND gm.user_id = $2
    AND gm.status = 'active' 
    AND gm.deleted_at IS NULL
    AND u.deleted_at IS NULL;

-- name: GetOverallUserBalance :many
-- Get user's balance across all groups
WITH payments AS (
    SELECT
        e.group_id,
        ep.user_id,
        SUM(ep.amount) AS total_paid
    FROM expense_payments ep
    JOIN expenses e ON e.id = ep.expense_id
    WHERE e.deleted_at IS NULL
      AND ep.deleted_at IS NULL
    GROUP BY e.group_id, ep.user_id
),
settlement_payments AS (
    SELECT
        s.group_id,
        s.payer_id AS user_id,
        SUM(s.amount) AS total_paid
    FROM settlements s
    WHERE s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
      AND s.group_id IS NOT NULL
      AND s.payer_id IS NOT NULL
    GROUP BY s.group_id, s.payer_id
),
splits AS (
    SELECT
        e.group_id,
        es.user_id,
        SUM(es.amount_owned) AS total_owed
    FROM expense_split es
    JOIN expenses e ON e.id = es.expense_id
    WHERE e.deleted_at IS NULL
      AND es.deleted_at IS NULL
    GROUP BY e.group_id, es.user_id
),
settlement_splits AS (
    SELECT
        s.group_id,
        s.payee_id AS user_id,
        SUM(s.amount) AS total_owed
    FROM settlements s
    WHERE s.type = 'group'
      AND s.status IN ('pending', 'completed')
      AND s.deleted_at IS NULL
      AND s.group_id IS NOT NULL
      AND s.payee_id IS NOT NULL
    GROUP BY s.group_id, s.payee_id
)
SELECT
    g.id AS group_id,
    g.name AS group_name,
    g.currency_code AS currency_code,
    (COALESCE(p.total_paid, 0::DECIMAL) + COALESCE(sp.total_paid, 0::DECIMAL))::TEXT AS total_paid,
    (COALESCE(s.total_owed, 0::DECIMAL) + COALESCE(ss.total_owed, 0::DECIMAL))::TEXT AS total_owed
FROM group_members gm
JOIN groups g ON gm.group_id = g.id
LEFT JOIN payments p ON p.group_id = gm.group_id AND p.user_id = gm.user_id
LEFT JOIN settlement_payments sp ON sp.group_id = gm.group_id AND sp.user_id = gm.user_id
LEFT JOIN splits s ON s.group_id = gm.group_id AND s.user_id = gm.user_id
LEFT JOIN settlement_splits ss ON ss.group_id = gm.group_id AND ss.user_id = gm.user_id
WHERE gm.user_id = $1
    AND gm.status = 'active'
    AND gm.deleted_at IS NULL
    AND g.deleted_at IS NULL
GROUP BY g.id, g.name, g.currency_code, p.total_paid, sp.total_paid, s.total_owed, ss.total_owed
ORDER BY g.name;
