# Database Schema Improvements - Split+

## Overview
This document outlines potential improvements to the Split+ database schema design, organized by category.

---

## 1. Data Integrity & Constraints

### Issues Identified
- Missing foreign key constraints (only references mentioned, not enforced)
- No check constraints for amounts (must be positive)
- Missing unique constraints where needed
- No validation for enum-like fields

### Recommendations

#### Foreign Keys
```sql
-- Add explicit foreign key constraints with CASCADE behavior
ALTER TABLE groups ADD CONSTRAINT fk_groups_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    
ALTER TABLE group_members ADD CONSTRAINT fk_group_members_group 
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;
    
ALTER TABLE group_members ADD CONSTRAINT fk_group_members_user 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    
-- Prevent duplicate memberships
ALTER TABLE group_members ADD CONSTRAINT uk_group_members_user_group 
    UNIQUE (group_id, user_id);
```

#### Check Constraints
```sql
-- Ensure amounts are positive
ALTER TABLE expenses ADD CONSTRAINT chk_expenses_amount_positive 
    CHECK (amount > 0);
    
ALTER TABLE expense_payments ADD CONSTRAINT chk_payments_amount_positive 
    CHECK (amount > 0);
    
ALTER TABLE expense_split ADD CONSTRAINT chk_split_amount_positive 
    CHECK (amount_owned >= 0);  -- Can be 0 if not participating
    
ALTER TABLE settlements ADD CONSTRAINT chk_settlements_amount_positive 
    CHECK (amount > 0);
    
-- Prevent self-settlements
ALTER TABLE settlements ADD CONSTRAINT chk_settlements_no_self_payment 
    CHECK (payer_id != payee_id);
```

#### Enum Constraints
```sql
-- Use CHECK constraints or ENUM types for role
ALTER TABLE group_members ADD CONSTRAINT chk_group_members_role 
    CHECK (role IN ('owner', 'admin', 'member'));

-- Split type validation
ALTER TABLE expense_split ADD CONSTRAINT chk_expense_split_type 
    CHECK (split_type IN ('equal', 'percentage', 'fixed', 'custom'));
```

---

## 2. Indexing Strategy

### Missing Indexes
- Foreign key columns (for JOIN performance)
- Frequently queried columns (email, group_id, expense_id)
- Soft delete queries (deleted_at IS NULL)
- Date range queries (date, paid_at, created_at)

### Recommendations
```sql
-- User lookups
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;

-- Group queries
CREATE INDEX idx_groups_created_by ON groups(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE UNIQUE INDEX idx_group_members_unique ON group_members(group_id, user_id) 
    WHERE deleted_at IS NULL;  -- If soft delete added

-- Expense queries
CREATE INDEX idx_expenses_group_id ON expenses(group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_created_by ON expenses(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_date ON expenses(date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_group_date ON expenses(group_id, date DESC) WHERE deleted_at IS NULL;

-- Expense relationships
CREATE INDEX idx_expense_payments_expense_id ON expense_payments(expense_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_payments_user_id ON expense_payments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_split_expense_id ON expense_split(expense_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_split_user_id ON expense_split(user_id) WHERE deleted_at IS NULL;

-- Settlement queries
CREATE INDEX idx_settlements_group_id ON settlements(group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_payer_id ON settlements(payer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_payee_id ON settlements(payee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_paid_at ON settlements(paid_at DESC) WHERE deleted_at IS NULL;
```

---

## 3. Schema Refinements

### 3.1 Users Table
**Current Issues:**
- Missing email verification
- No password reset tokens
- Language should use ISO 639-1 codes
- Missing timezone preference

**Suggestions:**
```sql
users
- id
- email (UNIQUE, indexed)
- email_verified_at (TIMESTAMPTZ)
- password_hash
- password_reset_token (TEXT, nullable, indexed)
- password_reset_expires_at (TIMESTAMPTZ)
- name
- avatar_url
- language (TEXT, DEFAULT 'en', CHECK IN ISO codes)
- timezone (TEXT, DEFAULT 'UTC')
- created_at
- updated_at
- deleted_at

-- Indexes
CREATE INDEX idx_users_email_verified ON users(email_verified_at) 
    WHERE email_verified_at IS NOT NULL;
CREATE INDEX idx_users_password_reset_token ON users(password_reset_token) 
    WHERE password_reset_token IS NOT NULL;
```

### 3.2 Groups Table
**Current Issues:**
- Metadata as JSON object (hard to query)
- Missing default currency
- No group settings

**Suggestions:**
```sql
groups
- id
- name
- description
- group_type (TEXT, CHECK IN ('trip', 'household', 'event', 'custom'))
- currency_code (TEXT, NOT NULL, DEFAULT 'USD')
- default_split_type (TEXT, DEFAULT 'equal')
- settings (JSONB) -- For flexible settings like notifications, etc.
- created_at
- created_by (FK to users.id)
- updated_at
- updated_by (FK to users.id)
- deleted_at

-- Indexes
CREATE INDEX idx_groups_type ON groups(group_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_groups_currency ON groups(currency_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_groups_settings ON groups USING GIN(settings); -- For JSONB queries
```

### 3.3 Group Members Table
**Current Issues:**
- Missing soft delete (should memberships be soft-deleted?)
- No invitation system
- Missing status (pending, active, inactive)

**Suggestions:**
```sql
group_members
- id
- group_id (FK to groups.id)
- user_id (FK to users.id)
- role (TEXT, CHECK IN ('owner', 'admin', 'member'), DEFAULT 'member')
- status (TEXT, CHECK IN ('pending', 'active', 'inactive'), DEFAULT 'pending')
- invited_by (FK to users.id, nullable)
- invited_at (TIMESTAMPTZ)
- joined_at (TIMESTAMPTZ)
- created_at
- updated_at
- deleted_at

-- Constraints
UNIQUE (group_id, user_id) WHERE deleted_at IS NULL
```

### 3.4 Expenses Table
**Current Issues:**
- Missing category/tags
- No receipt/image attachments
- Currency should match group currency (or allow multi-currency)
- Missing status (draft, pending, confirmed)

**Suggestions:**
```sql
expenses
- id
- group_id (FK to groups.id)
- title (TEXT, NOT NULL)
- description/notes (TEXT)
- amount (DECIMAL(19,4), NOT NULL, CHECK > 0)
- currency_code (TEXT, NOT NULL)
- exchange_rate (DECIMAL(19,8), nullable) -- If different from group currency
- category (TEXT, nullable) -- Food, Transport, etc.
- tags (TEXT[], nullable) -- Array of tags
- receipt_url (TEXT, nullable)
- date (DATE, NOT NULL, DEFAULT CURRENT_DATE)
- status (TEXT, CHECK IN ('draft', 'pending', 'confirmed'), DEFAULT 'confirmed')
- created_at
- created_by (FK to users.id)
- updated_at
- updated_by (FK to users.id)
- deleted_at

-- Indexes
CREATE INDEX idx_expenses_category ON expenses(category) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_tags ON expenses USING GIN(tags); -- For array searches
```

### 3.5 Expense Payments Table
**Current Issues:**
- Payment method should be enum
- Missing payment status
- No transaction reference

**Suggestions:**
```sql
expense_payments
- id
- expense_id (FK to expenses.id)
- user_id (FK to users.id)
- amount (DECIMAL(19,4), NOT NULL, CHECK > 0)
- currency_code (TEXT, NOT NULL) -- May differ from expense currency
- exchange_rate (DECIMAL(19,8), nullable)
- payment_method (TEXT, CHECK IN ('cash', 'card', 'bank_transfer', 'digital_wallet', 'other'))
- payment_status (TEXT, CHECK IN ('pending', 'completed', 'failed', 'refunded'), DEFAULT 'completed')
- transaction_reference (TEXT, nullable)
- paid_at (TIMESTAMPTZ, DEFAULT NOW())
- created_at
- updated_at
- deleted_at

-- Constraint: Sum of payments should not exceed expense amount (enforced in application)
```

### 3.6 Expense Split Table
**Current Issues:**
- `split_type` at split level is redundant (should be at expense level)
- `amount_owned` name is unclear (should be `amount_owed` or `share_amount`)
- Missing percentage/fixed value storage
- No validation that splits sum to expense amount

**Suggestions:**
```sql
expense_split
- id
- expense_id (FK to expenses.id)
- user_id (FK to users.id)
- share_amount (DECIMAL(19,4), NOT NULL, CHECK >= 0) -- Amount this user owes
- percentage (DECIMAL(5,2), nullable) -- If split_type is percentage
- fixed_amount (DECIMAL(19,4), nullable) -- If split_type is fixed
- is_paid (BOOLEAN, DEFAULT false) -- Quick check if user has paid their share
- created_at
- updated_at
- deleted_at

-- Indexes
CREATE INDEX idx_expense_split_expense_user ON expense_split(expense_id, user_id) 
    WHERE deleted_at IS NULL;

-- Note: Split type should be moved to expenses table
-- Validation: Sum of share_amount should equal expense.amount (enforced in application)
```

### 3.7 Settlements Table
**Current Issues:**
- Missing status (pending, completed, cancelled)
- No payment method
- Missing transaction reference
- Should link to expense_payments if applicable

**Suggestions:**
```sql
settlements
- id
- group_id (FK to groups.id)
- payer_id (FK to users.id)
- payee_id (FK to users.id)
- amount (DECIMAL(19,4), NOT NULL, CHECK > 0)
- currency_code (TEXT, NOT NULL)
- status (TEXT, CHECK IN ('pending', 'completed', 'cancelled'), DEFAULT 'pending')
- payment_method (TEXT, CHECK IN ('cash', 'card', 'bank_transfer', 'digital_wallet', 'other'), nullable)
- transaction_reference (TEXT, nullable)
- paid_at (TIMESTAMPTZ, nullable) -- NULL if pending
- notes (TEXT)
- created_at
- created_by (FK to users.id) -- Who recorded this settlement
- updated_at
- updated_by (FK to users.id)
- deleted_at

-- Constraint
CHECK (payer_id != payee_id)
```

---

## 4. Data Consistency & Business Rules

### 4.1 Split Validation
**Issue:** Need to ensure expense splits sum to expense amount

**Solution:** Application-level validation with database triggers or application logic
```sql
-- Function to validate split totals
CREATE OR REPLACE FUNCTION validate_expense_split()
RETURNS TRIGGER AS $$
DECLARE
    expense_amount DECIMAL(19,4);
    split_total DECIMAL(19,4);
BEGIN
    SELECT amount INTO expense_amount FROM expenses WHERE id = NEW.expense_id;
    SELECT COALESCE(SUM(share_amount), 0) INTO split_total 
    FROM expense_split 
    WHERE expense_id = NEW.expense_id AND deleted_at IS NULL;
    
    IF split_total > expense_amount THEN
        RAISE EXCEPTION 'Split total (%) exceeds expense amount (%)', split_total, expense_amount;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_validate_expense_split
    AFTER INSERT OR UPDATE ON expense_split
    FOR EACH ROW EXECUTE FUNCTION validate_expense_split();
```

### 4.2 Payment Validation
**Issue:** Payments should not exceed expense amount

**Solution:** Similar trigger or application logic

### 4.3 Currency Consistency
**Issue:** Multiple currency fields need consistency

**Suggestions:**
- Store base currency at group level
- Use exchange_rate for conversions
- Consider a `currency_exchanges` table for historical rates
- Or use a currency service/API

---

## 5. Additional Tables to Consider

### 5.1 Currency Exchange Rates
```sql
currency_exchanges
- id
- from_currency (TEXT)
- to_currency (TEXT)
- rate (DECIMAL(19,8))
- effective_date (DATE)
- created_at
- UNIQUE (from_currency, to_currency, effective_date)
```

### 5.2 Notifications
```sql
notifications
- id
- user_id (FK to users.id)
- type (TEXT) -- expense_added, settlement_requested, etc.
- title (TEXT)
- message (TEXT)
- metadata (JSONB)
- read_at (TIMESTAMPTZ)
- created_at
- INDEX (user_id, read_at, created_at DESC)
```

### 5.3 Activity Log / Audit Trail
```sql
activity_log
- id
- user_id (FK to users.id, nullable)
- group_id (FK to groups.id, nullable)
- action (TEXT) -- created_expense, updated_settlement, etc.
- entity_type (TEXT) -- expense, settlement, group, etc.
- entity_id (BIGINT)
- changes (JSONB) -- Before/after state
- ip_address (INET, nullable)
- user_agent (TEXT, nullable)
- created_at
- INDEX (group_id, created_at DESC)
- INDEX (user_id, created_at DESC)
```

### 5.4 Group Invitations
```sql
group_invitations
- id
- group_id (FK to groups.id)
- email (TEXT) -- For non-users
- user_id (FK to users.id, nullable) -- If user exists
- invited_by (FK to users.id)
- token (TEXT, UNIQUE) -- For email links
- status (TEXT, CHECK IN ('pending', 'accepted', 'declined', 'expired'))
- expires_at (TIMESTAMPTZ)
- created_at
- accepted_at (TIMESTAMPTZ, nullable)
```

---

## 6. Performance Optimizations

### 6.1 Materialized Views
For frequently accessed aggregated data:
```sql
-- Group balance summary
CREATE MATERIALIZED VIEW group_balances AS
SELECT 
    g.id as group_id,
    u.id as user_id,
    COALESCE(SUM(es.share_amount), 0) as total_owed,
    COALESCE(SUM(ep.amount), 0) as total_paid,
    COALESCE(SUM(ep.amount), 0) - COALESCE(SUM(es.share_amount), 0) as balance
FROM groups g
CROSS JOIN group_members gm
JOIN users u ON gm.user_id = u.id
LEFT JOIN expenses e ON e.group_id = g.id AND e.deleted_at IS NULL
LEFT JOIN expense_split es ON es.expense_id = e.id AND es.user_id = u.id AND es.deleted_at IS NULL
LEFT JOIN expense_payments ep ON ep.expense_id = e.id AND ep.user_id = u.id AND ep.deleted_at IS NULL
WHERE g.deleted_at IS NULL AND gm.deleted_at IS NULL
GROUP BY g.id, u.id;

CREATE UNIQUE INDEX ON group_balances(group_id, user_id);
-- Refresh periodically or on expense/settlement changes
```

### 6.2 Partial Indexes
For soft-deleted records:
```sql
-- Only index non-deleted records
CREATE INDEX idx_expenses_active ON expenses(group_id, date DESC) 
    WHERE deleted_at IS NULL;
```

---

## 7. Security Considerations

### 7.1 Row-Level Security (RLS)
If using PostgreSQL, consider RLS policies:
```sql
-- Users can only see their own data
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_isolation ON users
    FOR ALL USING (id = current_setting('app.user_id')::BIGINT);

-- Group members can only see their groups
ALTER TABLE groups ENABLE ROW LEVEL SECURITY;
CREATE POLICY group_member_access ON groups
    FOR SELECT USING (
        id IN (
            SELECT group_id FROM group_members 
            WHERE user_id = current_setting('app.user_id')::BIGINT
        )
    );
```

### 7.2 Sensitive Data
- Consider encrypting `password_reset_token`
- Hash `password_hash` properly (bcrypt/argon2)
- Consider encrypting payment method details if storing sensitive info

---

## 8. Migration Strategy

### 8.1 Versioning
- Use semantic versioning for schema changes
- Document breaking changes
- Provide migration scripts for data transformations

### 8.2 Backward Compatibility
- Add new columns as nullable initially
- Use feature flags for new functionality
- Gradual rollout of constraints

---

## 9. Query Optimization Patterns

### 9.1 Common Queries to Optimize
1. **User's balance in a group** - Use materialized view or computed column
2. **Recent expenses in group** - Index on (group_id, date DESC)
3. **Unpaid splits** - Index on (user_id, is_paid) WHERE is_paid = false
4. **Pending settlements** - Index on (group_id, status) WHERE status = 'pending'

---

## 10. Summary of Key Improvements

### High Priority
1. ✅ Add foreign key constraints
2. ✅ Add check constraints for amounts and enums
3. ✅ Add indexes on foreign keys and frequently queried columns
4. ✅ Add unique constraint on (group_id, user_id) in group_members
5. ✅ Move `split_type` from expense_split to expenses table
6. ✅ Add status fields where needed (expenses, settlements, group_members)
7. ✅ Add soft delete to group_members

### Medium Priority
8. Add email verification and password reset to users
9. Add currency exchange rate handling
10. Add notifications table
11. Add activity log/audit trail
12. Add group invitations table
13. Add category and tags to expenses

### Low Priority
14. Add materialized views for aggregated data
15. Consider row-level security
16. Add receipt/image storage fields
17. Add timezone support

---

## Questions to Consider

1. **Multi-currency support**: How complex should this be? Real-time rates or manual entry?
2. **Split types**: Should custom splits be supported (beyond equal/percentage/fixed)?
3. **Payment tracking**: Should payments be linked to specific splits or just expenses?
4. **Settlement automation**: Should the system suggest optimal settlements?
5. **Data retention**: How long to keep soft-deleted records?
6. **Archiving**: Should old expenses/groups be archived separately?
7. **Receipt storage**: Store URLs or actual files? Use S3/cloud storage?
