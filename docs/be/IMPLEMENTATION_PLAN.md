# Split+ Feature Implementation Plan

## Selected Features to Implement

Based on MVP priorities, we'll implement:
1. **Balance Calculations** - Core feature to show who owes whom
2. **Settlements** - Record payments between users
3. **Recurring Expenses** - Automate repeated expenses
4. **Comments on Expenses** - User interaction and context

---

## Feature 1: Balance Calculations

### Overview
Calculate and display balances showing who owes whom in each group. This is the core value proposition of a bill-splitting app.

### Implementation Steps

#### 1.1 Database Queries
**File**: `internal/db/queries/balances.sql`

Create SQL queries to:
- Calculate per-user balances in a group (total paid - total owed)
- Calculate who owes whom (simplified debt graph)
- Get overall balance for a user across all groups

**Key Query Logic**:
```sql
-- Balance per user in a group
SELECT 
    u.id,
    u.email,
    u.name,
    COALESCE(SUM(ep.amount), 0) as total_paid,
    COALESCE(SUM(es.amount_owned), 0) as total_owed,
    COALESCE(SUM(ep.amount), 0) - COALESCE(SUM(es.amount_owned), 0) as balance
FROM group_members gm
JOIN users u ON gm.user_id = u.id
LEFT JOIN expenses e ON e.group_id = gm.group_id AND e.deleted_at IS NULL
LEFT JOIN expense_payments ep ON ep.expense_id = e.id AND ep.user_id = u.id AND ep.deleted_at IS NULL
LEFT JOIN expense_split es ON es.expense_id = e.id AND es.user_id = u.id AND es.deleted_at IS NULL
WHERE gm.group_id = $1 AND gm.status = 'active' AND gm.deleted_at IS NULL
GROUP BY u.id, u.email, u.name
```

#### 1.2 Repository Layer
**File**: `internal/repository/balance_repository.go`

- Interface: `BalanceRepository`
- Methods:
  - `GetGroupBalances(ctx, groupID) ([]BalanceRow, error)`
  - `GetUserBalanceInGroup(ctx, groupID, userID) (Balance, error)`
  - `GetOverallUserBalance(ctx, userID) ([]GroupBalance, error)`

#### 1.3 Service Layer
**File**: `internal/service/balance_service.go`

- Interface: `BalanceService`
- Methods:
  - `GetGroupBalances(ctx, groupID, requesterID) ([]BalanceResponse, error)`
  - `GetUserBalanceInGroup(ctx, groupID, userID, requesterID) (BalanceResponse, error)`
  - `GetOverallBalance(ctx, userID) (OverallBalanceResponse, error)`
  - `GetSimplifiedDebts(ctx, groupID, requesterID) ([]Debt, error)` - who owes whom

**Business Logic**:
- Validate requester is group member
- Calculate balances from expenses (payments - splits)
- Handle currency consistency
- Return simplified debt graph (A owes B $X)

#### 1.4 Handler Layer
**File**: `internal/http/handlers/balances.go`

**Endpoints**:
- `GET /groups/{group_id}/balances` - List all balances in group
- `GET /groups/{group_id}/balances/{user_id}` - Get specific user's balance
- `GET /users/me/balances` - Get user's balances across all groups
- `GET /groups/{group_id}/debts` - Get simplified "who owes whom" view

#### 1.5 Router
**File**: `internal/http/router/balance_router.go`

Add `WithBalanceRoutes(balanceService)` to router options.

#### 1.6 Wire Up
**File**: `internal/app/app.go`

Add balance repository, service, and routes.

---

## Feature 2: Settlements

### Overview
Allow users to record payments made to settle debts. This is how users actually pay each other back.

### Implementation Steps

#### 2.1 Database Migration
**File**: `internal/db/migrations/YYYYMMDDHHMMSS_create_settlements.sql`

Create settlements table:
```sql
CREATE TABLE settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id),
    payer_id UUID NOT NULL REFERENCES users(id),
    payee_id UUID NOT NULL REFERENCES users(id),
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    currency_code TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled')),
    payment_method TEXT,
    transaction_reference TEXT,
    paid_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    CHECK (payer_id != payee_id)
);
```

#### 2.2 Database Queries
**File**: `internal/db/queries/settlements.sql`

- `CreateSettlement`
- `GetSettlementByID`
- `ListSettlementsByGroup`
- `ListSettlementsByUser`
- `UpdateSettlementStatus`
- `UpdateSettlement`

#### 2.3 Repository Layer
**File**: `internal/repository/settlement_repository.go`

- Interface: `SettlementRepository`
- Methods matching SQL queries

#### 2.4 Service Layer
**File**: `internal/service/settlement_service.go`

- Interface: `SettlementService`
- Methods:
  - `CreateSettlement(ctx, input) (Settlement, error)`
  - `GetSettlementByID(ctx, settlementID, requesterID) (Settlement, error)`
  - `ListSettlementsByGroup(ctx, groupID, requesterID) ([]Settlement, error)`
  - `UpdateSettlementStatus(ctx, settlementID, status, requesterID) (Settlement, error)`
  - `UpdateSettlement(ctx, input) (Settlement, error)`
  - `DeleteSettlement(ctx, settlementID, requesterID) error`

**Business Logic**:
- Validate payer/payee are group members
- Validate amount > 0
- Validate payer != payee
- When marking as completed, set `paid_at` timestamp
- Validate requester permissions

#### 2.5 Handler Layer
**File**: `internal/http/handlers/settlements.go`

**Endpoints**:
- `POST /groups/{group_id}/settlements` - Create settlement
- `GET /groups/{group_id}/settlements` - List settlements in group
- `GET /groups/{group_id}/settlements/{settlement_id}` - Get settlement
- `PUT /groups/{group_id}/settlements/{settlement_id}` - Update settlement
- `PATCH /groups/{group_id}/settlements/{settlement_id}/status` - Update status
- `DELETE /groups/{group_id}/settlements/{settlement_id}` - Delete settlement

#### 2.6 Router
**File**: `internal/http/router/settlement_router.go`

Add `WithSettlementRoutes(settlementService)` to router.

#### 2.7 Wire Up
**File**: `internal/app/app.go`

Add settlement repository, service, and routes.

---

## Feature 3: Recurring Expenses

### Overview
Allow users to create expense templates that automatically generate expenses on a schedule (weekly, monthly, yearly).

### Implementation Steps

#### 3.1 Database Migration
**File**: `internal/db/migrations/YYYYMMDDHHMMSS_create_recurring_expenses.sql`

Add fields to expenses table or create separate table:
```sql
-- Option 1: Add columns to expenses table
ALTER TABLE expenses ADD COLUMN is_recurring BOOLEAN DEFAULT false;
ALTER TABLE expenses ADD COLUMN recurring_template_id UUID REFERENCES expenses(id);
ALTER TABLE expenses ADD COLUMN repeat_interval TEXT CHECK (repeat_interval IN ('weekly', 'monthly', 'yearly'));
ALTER TABLE expenses ADD COLUMN next_occurrence_date DATE;
ALTER TABLE expenses ADD COLUMN end_date DATE;

-- Option 2: Separate recurring_expenses table (better for complex cases)
CREATE TABLE recurring_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id),
    title TEXT NOT NULL,
    notes TEXT,
    amount DECIMAL(10, 2) NOT NULL,
    currency_code TEXT NOT NULL,
    repeat_interval TEXT NOT NULL CHECK (repeat_interval IN ('weekly', 'monthly', 'yearly')),
    start_date DATE NOT NULL,
    end_date DATE,
    next_occurrence_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL REFERENCES users(id),
    deleted_at TIMESTAMPTZ
);
```

**Recommendation**: Start with Option 1 (simpler), can refactor to Option 2 later if needed.

#### 3.2 Database Queries
**File**: `internal/db/queries/recurring_expenses.sql`

- `CreateRecurringExpense`
- `GetRecurringExpenseByID`
- `ListRecurringExpensesByGroup`
- `UpdateRecurringExpense`
- `DeleteRecurringExpense`
- `GetRecurringExpensesDue` - for cron job

#### 3.3 Repository Layer
**File**: `internal/repository/recurring_expense_repository.go`

- Interface: `RecurringExpenseRepository`
- Methods matching SQL queries

#### 3.4 Service Layer
**File**: `internal/service/recurring_expense_service.go`

- Interface: `RecurringExpenseService`
- Methods:
  - `CreateRecurringExpense(ctx, input) (RecurringExpense, error)`
  - `GetRecurringExpenseByID(ctx, id, requesterID) (RecurringExpense, error)`
  - `ListRecurringExpensesByGroup(ctx, groupID, requesterID) ([]RecurringExpense, error)`
  - `UpdateRecurringExpense(ctx, input) (RecurringExpense, error)`
  - `DeleteRecurringExpense(ctx, id, requesterID) error`
  - `GenerateExpenseFromRecurring(ctx, recurringID, requesterID) (Expense, error)` - Manual trigger

**Business Logic**:
- Validate group membership
- Calculate next occurrence date based on interval
- Validate dates (start_date <= end_date if provided)
- Handle timezone considerations

#### 3.5 Handler Layer
**File**: `internal/http/handlers/recurring_expenses.go`

**Endpoints**:
- `POST /groups/{group_id}/recurring-expenses` - Create recurring expense
- `GET /groups/{group_id}/recurring-expenses` - List recurring expenses
- `GET /groups/{group_id}/recurring-expenses/{id}` - Get recurring expense
- `PUT /groups/{group_id}/recurring-expenses/{id}` - Update recurring expense
- `DELETE /groups/{group_id}/recurring-expenses/{id}` - Delete recurring expense
- `POST /groups/{group_id}/recurring-expenses/{id}/generate` - Manually generate expense

#### 3.6 Router
**File**: `internal/http/router/recurring_expense_router.go`

Add `WithRecurringExpenseRoutes(recurringExpenseService)` to router.

#### 3.7 Background Job (Future)
**File**: `internal/job/recurring_expense_generator.go` (optional, for Phase 2)

Cron job to automatically generate expenses from recurring templates.

#### 3.8 Wire Up
**File**: `internal/app/app.go`

Add recurring expense repository, service, and routes.

---

## Feature 4: Comments on Expenses

### Overview
Allow users to add comments to expenses for context, questions, or discussions.

### Implementation Steps

#### 4.1 Database Migration
**File**: `internal/db/migrations/YYYYMMDDHHMMSS_create_expense_comments.sql`

```sql
CREATE TABLE expense_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_expense_comments_expense_id ON expense_comments(expense_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_comments_user_id ON expense_comments(user_id) WHERE deleted_at IS NULL;
```

#### 4.2 Database Queries
**File**: `internal/db/queries/expense_comments.sql`

- `CreateExpenseComment`
- `GetExpenseCommentByID`
- `ListExpenseComments` - with user info joined
- `UpdateExpenseComment`
- `DeleteExpenseComment`

#### 4.3 Repository Layer
**File**: `internal/repository/expense_comment_repository.go`

- Interface: `ExpenseCommentRepository`
- Methods matching SQL queries

#### 4.4 Service Layer
**File**: `internal/service/expense_comment_service.go`

- Interface: `ExpenseCommentService`
- Methods:
  - `CreateComment(ctx, input) (Comment, error)`
  - `GetCommentByID(ctx, commentID, requesterID) (Comment, error)`
  - `ListCommentsByExpense(ctx, expenseID, requesterID) ([]Comment, error)`
  - `UpdateComment(ctx, input) (Comment, error)`
  - `DeleteComment(ctx, commentID, requesterID) error`

**Business Logic**:
- Validate requester is group member (via expense)
- Validate comment text is not empty
- Only allow comment author or expense creator to edit/delete
- Return comments with user info (email, name, avatar)

#### 4.5 Handler Layer
**File**: `internal/http/handlers/expense_comments.go`

**Endpoints**:
- `POST /groups/{group_id}/expenses/{expense_id}/comments` - Add comment
- `GET /groups/{group_id}/expenses/{expense_id}/comments` - List comments
- `GET /groups/{group_id}/expenses/{expense_id}/comments/{comment_id}` - Get comment
- `PUT /groups/{group_id}/expenses/{expense_id}/comments/{comment_id}` - Update comment
- `DELETE /groups/{group_id}/expenses/{expense_id}/comments/{comment_id}` - Delete comment

#### 4.6 Router
**File**: `internal/http/router/expense_comment_router.go`

Add `WithExpenseCommentRoutes(expenseCommentService)` to router.

#### 4.7 Wire Up
**File**: `internal/app/app.go`

Add expense comment repository, service, and routes.

---

## Implementation Order & Dependencies

### Phase 1: Core Functionality (Week 1-2)
1. **Balance Calculations** - Foundation for everything else
   - Dependencies: None (uses existing expenses)
   - Priority: HIGHEST

2. **Settlements** - Depends on understanding balances
   - Dependencies: Balance calculations (conceptually)
   - Priority: HIGH

### Phase 2: Enhanced Features (Week 3-4)
3. **Recurring Expenses** - Standalone feature
   - Dependencies: None
   - Priority: MEDIUM

4. **Comments on Expenses** - Standalone feature
   - Dependencies: None
   - Priority: MEDIUM

---

## Testing Strategy

For each feature:
1. **Service Layer Tests** (`*_service_test.go`)
   - Test business logic
   - Test validation
   - Test error cases
   - Test edge cases

2. **Handler Layer Tests** (`*_test.go`)
   - Test HTTP endpoints
   - Test request/response formats
   - Test authentication/authorization
   - Test error handling

3. **Integration Tests** (optional)
   - Test full flow from API to database
   - Test complex scenarios

---

## Estimated Effort

- **Balance Calculations**: 2-3 days
- **Settlements**: 2-3 days
- **Recurring Expenses**: 3-4 days
- **Comments**: 1-2 days

**Total**: ~8-12 days of focused development

---

## Notes & Considerations

### Security Note
⚠️ **Important**: Authentication is currently using `X-User-ID` header (placeholder). For production, implement proper JWT authentication before deploying these features.

### Balance Calculation Algorithm
The balance calculation needs to:
1. Sum all payments made by a user (from `expense_payments`)
2. Sum all amounts owed by a user (from `expense_split`)
3. Calculate: `balance = total_paid - total_owed`
4. Positive balance = user is owed money
5. Negative balance = user owes money

### Settlement Integration
Settlements should ideally update balances, but for MVP we can:
- Calculate balances from expenses only (settlements are separate records)
- Or: Include settlements in balance calculations (subtract from debts)

**Recommendation**: Start simple - settlements are records of payments, balances calculated from expenses. Can enhance later.

### Recurring Expenses
For MVP:
- Manual generation only (user clicks "generate")
- Background job for auto-generation can be Phase 2

### Comments
Keep it simple:
- No nested replies (flat comments only)
- No mentions/notifications (Phase 2)
- Basic CRUD operations

---

## Next Steps

1. Review and approve this plan
2. Start with **Balance Calculations** (highest priority)
3. Follow the implementation order
4. Test each feature before moving to next
5. Update Bruno API collection as we go
