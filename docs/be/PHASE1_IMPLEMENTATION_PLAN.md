# Phase 1 Implementation Plan - Split+ Backend

## Executive Summary

This plan outlines the implementation strategy for **8 core features** to be delivered in Phase 1. Features are prioritized based on user impact and implementation dependencies.

**Timeline:** 6-8 weeks  
**Features:** 8 major features  
**Approach:** Feature-by-feature implementation with testing

---

## Feature Priority Matrix

| Priority | Feature | User Impact | Complexity | Duration |
|----------|---------|-------------|------------|----------|
| 🔴 P0 | Expense Categories | High | Low | 3-4 days |
| 🔴 P0 | Comments & Activity Feed | High | Medium | 5-6 days |
| 🟡 P1 | Search & Filtering | High | Medium | 4-5 days |
| 🟡 P1 | Group Invitations (Email) | Medium | Medium | 4-5 days |
| 🟢 P2 | Group Settings & Customization | Medium | Low | 2-3 days |
| 🟢 P2 | Expense Simplification | Medium | Medium | 3-4 days |
| 🟢 P2 | Group Notes (Whiteboard) | Medium | Low | 2-3 days |
| 🟢 P2 | Group Archiving | Low | Low | 2 days |

**Total Estimated Time:** 27-36 days (~6-8 weeks)

---

## Implementation Order

### Week 1-2: Foundation Features
1. **Expense Categories** (P0)
2. **Group Settings & Customization** (P2)

### Week 3-4: Collaboration Features
3. **Comments & Activity Feed** (P0)
4. **Group Notes** (P2)

### Week 5-6: User Experience Features
5. **Search & Filtering** (P1)
6. **Group Archiving** (P2)

### Week 7-8: Advanced Features
7. **Group Invitations via Email** (P1)
8. **Expense Simplification** (P2)

---

# Feature 1: Expense Categories & Organization

**Priority:** 🔴 P0  
**Duration:** 3-4 days  
**Dependencies:** None

## Overview
Add category support to expenses for better organization and reporting. Categories help users understand spending patterns.

## Database Changes

### Migration: `create_expense_categories.sql`
```sql
-- +goose Up
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    icon TEXT, -- emoji or icon identifier
    color TEXT, -- hex color code
    is_system BOOLEAN DEFAULT false, -- system vs user-defined
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- System categories
INSERT INTO expense_categories (name, icon, color, is_system) VALUES
    ('Food & Drink', '🍔', '#FF6B6B', true),
    ('Entertainment', '🎬', '#4ECDC4', true),
    ('Home', '🏠', '#45B7D1', true),
    ('Transportation', '🚗', '#96CEB4', true),
    ('Shopping', '🛍️', '#FFEAA7', true),
    ('Utilities', '💡', '#DFE6E9', true),
    ('Other', '📌', '#B2BEC3', true);

CREATE INDEX idx_expense_categories_name ON expense_categories(name) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS expense_categories;
```

### Migration: `add_category_to_expenses.sql`
```sql
-- +goose Up
ALTER TABLE expenses ADD COLUMN category_id UUID REFERENCES expense_categories(id);
ALTER TABLE expenses ADD COLUMN tags TEXT[] DEFAULT '{}';

CREATE INDEX idx_expenses_category_id ON expenses(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_tags ON expenses USING GIN(tags);

-- +goose Down
ALTER TABLE expenses DROP COLUMN IF EXISTS category_id;
ALTER TABLE expenses DROP COLUMN IF EXISTS tags;
```

## Implementation Steps

### 1. Database Queries (`internal/db/queries/expense_categories.sql`)
```sql
-- name: CreateExpenseCategory :one
-- name: GetExpenseCategoryByID :one
-- name: ListExpenseCategories :many
-- name: UpdateExpenseCategory :one
-- name: DeleteExpenseCategory :exec
-- name: GetSystemCategories :many
```

### 2. Repository Layer (`internal/repository/expense_category_repository.go`)
```go
type ExpenseCategoryRepository interface {
    CreateCategory(ctx context.Context, params sqlc.CreateExpenseCategoryParams) (sqlc.ExpenseCategory, error)
    GetCategoryByID(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error)
    ListCategories(ctx context.Context) ([]sqlc.ExpenseCategory, error)
    GetSystemCategories(ctx context.Context) ([]sqlc.ExpenseCategory, error)
    UpdateCategory(ctx context.Context, params sqlc.UpdateExpenseCategoryParams) (sqlc.ExpenseCategory, error)
    DeleteCategory(ctx context.Context, id pgtype.UUID) error
}
```

### 3. Service Layer (`internal/service/expense_category_service.go`)
- Validate category names
- Prevent deletion of system categories
- Handle category assignment to expenses
- Update expense service to include category_id

### 4. API Endpoints
```
GET    /categories                    - List all categories
GET    /categories/system             - List system categories
POST   /categories                    - Create custom category
PUT    /categories/{id}               - Update category
DELETE /categories/{id}               - Delete category

# Update expense endpoints to include category_id
POST   /groups/{group_id}/expenses    - Include category_id
PUT    /groups/{group_id}/expenses/{id} - Include category_id
```

### 5. Testing
- Service tests for category CRUD
- Test category assignment to expenses
- Test system category protection
- Integration tests for expense creation with categories

---

# Feature 2: Comments & Activity Feed

**Priority:** 🔴 P0  
**Duration:** 5-6 days  
**Dependencies:** None

## Overview
Enable collaboration through expense comments and track all group activities. Activity feed shows expense history using the same infrastructure.

## Database Changes

### Migration: `create_expense_comments.sql`
```sql
-- +goose Up
CREATE TABLE expense_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL CHECK (length(comment) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_expense_comments_expense_id ON expense_comments(expense_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_comments_user_id ON expense_comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_comments_created_at ON expense_comments(created_at DESC) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS expense_comments;
```

### Migration: `create_activity_feed.sql`
```sql
-- +goose Up
CREATE TABLE activity_feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id), -- actor
    action TEXT NOT NULL, -- expense_created, expense_updated, expense_deleted, settlement_created, etc.
    entity_type TEXT NOT NULL, -- expense, settlement, group, member
    entity_id UUID NOT NULL,
    metadata JSONB DEFAULT '{}', -- flexible data storage
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_activity_feed_group_id ON activity_feed(group_id, created_at DESC);
CREATE INDEX idx_activity_feed_user_id ON activity_feed(user_id, created_at DESC);
CREATE INDEX idx_activity_feed_entity ON activity_feed(entity_type, entity_id);
CREATE INDEX idx_activity_feed_metadata ON activity_feed USING GIN(metadata);

-- +goose Down
DROP TABLE IF EXISTS activity_feed;
```

## Implementation Steps

### 1. Database Queries
**`expense_comments.sql`:**
```sql
-- name: CreateExpenseComment :one
-- name: GetExpenseCommentByID :one
-- name: ListExpenseComments :many (with user info joined)
-- name: UpdateExpenseComment :one
-- name: DeleteExpenseComment :exec
-- name: CountExpenseComments :one
```

**`activity_feed.sql`:**
```sql
-- name: CreateActivity :one
-- name: ListGroupActivities :many
-- name: ListUserActivities :many
-- name: GetExpenseHistory :many (activities for specific expense)
```

### 2. Repository Layer
- `ExpenseCommentRepository`
- `ActivityFeedRepository`

### 3. Service Layer
**`expense_comment_service.go`:**
- Validate user is group member
- Create activity when comment added
- Only allow author to edit/delete

**`activity_feed_service.go`:**
- Helper methods to create activities
- Format activity messages
- Pagination support

### 4. Integration with Existing Services
Update these services to create activity feed entries:
- `expense_service.go` - expense_created, expense_updated, expense_deleted
- `settlement_service.go` - settlement_created, settlement_completed
- `group_service.go` - member_added, member_joined, member_removed

### 5. API Endpoints
```
# Comments
POST   /groups/{group_id}/expenses/{expense_id}/comments
GET    /groups/{group_id}/expenses/{expense_id}/comments
PUT    /groups/{group_id}/expenses/{expense_id}/comments/{id}
DELETE /groups/{group_id}/expenses/{expense_id}/comments/{id}

# Activity Feed
GET    /groups/{group_id}/activity
GET    /users/me/activity
GET    /groups/{group_id}/expenses/{expense_id}/history
```

### 6. Testing
- Comment CRUD tests
- Activity creation tests
- Permission tests (only group members can comment)
- Integration tests for activity feed

---

# Feature 3: Search & Filtering

**Priority:** 🟡 P1  
**Duration:** 4-5 days  
**Dependencies:** Expense Categories (optional)

## Overview
Enable users to search expenses by keyword and filter by date range, category, person, and amount.

## Database Changes

### Migration: `add_search_indexes.sql`
```sql
-- +goose Up
-- Full-text search on expense title and notes
CREATE INDEX idx_expenses_title_search ON expenses USING GIN(to_tsvector('english', title));
CREATE INDEX idx_expenses_notes_search ON expenses USING GIN(to_tsvector('english', notes)) WHERE notes IS NOT NULL;

-- Composite indexes for common filter combinations
CREATE INDEX idx_expenses_group_date_category ON expenses(group_id, date DESC, category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_amount_range ON expenses(group_id, amount) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_expenses_title_search;
DROP INDEX IF EXISTS idx_expenses_notes_search;
DROP INDEX IF EXISTS idx_expenses_group_date_category;
DROP INDEX IF EXISTS idx_expenses_amount_range;
```

## Implementation Steps

### 1. Database Queries (`expenses.sql` - add new queries)
```sql
-- name: SearchExpenses :many
SELECT e.* FROM expenses e
WHERE e.group_id = $1
  AND e.deleted_at IS NULL
  AND (
    -- Text search
    ($2::text IS NULL OR to_tsvector('english', e.title || ' ' || COALESCE(e.notes, '')) @@ plainto_tsquery('english', $2))
  )
  AND ($3::date IS NULL OR e.date >= $3) -- start_date
  AND ($4::date IS NULL OR e.date <= $4) -- end_date
  AND ($5::uuid IS NULL OR e.category_id = $5)
  AND ($6::uuid IS NULL OR e.created_by = $6)
  AND ($7::decimal IS NULL OR e.amount >= $7) -- min_amount
  AND ($8::decimal IS NULL OR e.amount <= $8) -- max_amount
ORDER BY e.date DESC, e.created_at DESC
LIMIT $9 OFFSET $10;

-- name: SearchExpensesByPayer :many
-- Search expenses where specific user paid
SELECT DISTINCT e.* FROM expenses e
JOIN expense_payments ep ON ep.expense_id = e.id
WHERE e.group_id = $1
  AND ep.user_id = $2
  AND e.deleted_at IS NULL
  AND ep.deleted_at IS NULL
ORDER BY e.date DESC;

-- name: SearchExpensesByOwer :many
-- Search expenses where specific user owes
SELECT DISTINCT e.* FROM expenses e
JOIN expense_split es ON es.expense_id = e.id
WHERE e.group_id = $1
  AND es.user_id = $2
  AND e.deleted_at IS NULL
  AND es.deleted_at IS NULL
ORDER BY e.date DESC;
```

### 2. Service Layer (`expense_service.go` - add methods)
```go
type SearchExpensesInput struct {
    GroupID      pgtype.UUID
    Query        string      // text search
    StartDate    *time.Time
    EndDate      *time.Time
    CategoryID   *pgtype.UUID
    UserID       *pgtype.UUID // created by
    MinAmount    *string
    MaxAmount    *string
    PayerID      *pgtype.UUID // who paid
    OwerID       *pgtype.UUID // who owes
    Limit        int32
    Offset       int32
}

func (s *expenseService) SearchExpenses(ctx context.Context, input SearchExpensesInput) ([]sqlc.Expense, error)
```

### 3. API Endpoints
```
GET /groups/{group_id}/expenses/search?q=dinner&start_date=2024-01-01&end_date=2024-12-31&category_id=xxx&user_id=xxx&min_amount=10&max_amount=100&payer_id=xxx&ower_id=xxx&limit=20&offset=0
```

### 4. Testing
- Test text search functionality
- Test date range filtering
- Test category filtering
- Test amount range filtering
- Test user filtering (payer/ower)
- Test pagination
- Test combined filters

---

# Feature 4: Group Invitations via Email

**Priority:** 🟡 P1  
**Duration:** 4-5 days  
**Dependencies:** None

## Overview
Allow inviting non-users via email. Support splitting expenses with non-joined users (pending members).

## Database Changes

### Migration: `create_group_invitations.sql`
```sql
-- +goose Up
CREATE TABLE group_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    user_id UUID REFERENCES users(id), -- NULL if user doesn't exist yet
    invited_by UUID NOT NULL REFERENCES users(id),
    token TEXT NOT NULL UNIQUE, -- for email link
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_group_invitations_group_id ON group_invitations(group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_group_invitations_email ON group_invitations(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_group_invitations_token ON group_invitations(token) WHERE deleted_at IS NULL;
CREATE INDEX idx_group_invitations_status ON group_invitations(status) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS group_invitations;
```

### Migration: `create_pending_users.sql`
```sql
-- +goose Up
-- Temporary users for non-joined members
CREATE TABLE pending_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    name TEXT, -- optional display name
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    converted_user_id UUID REFERENCES users(id), -- when they sign up
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_pending_users_email ON pending_users(email) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS pending_users;
```

### Migration: `update_expense_references.sql`
```sql
-- +goose Up
-- Allow expense_payments and expense_split to reference pending users
ALTER TABLE expense_payments ADD COLUMN pending_user_id UUID REFERENCES pending_users(id);
ALTER TABLE expense_split ADD COLUMN pending_user_id UUID REFERENCES pending_users(id);

-- Add constraint: either user_id or pending_user_id must be set
ALTER TABLE expense_payments ADD CONSTRAINT chk_expense_payments_user_or_pending 
    CHECK ((user_id IS NOT NULL AND pending_user_id IS NULL) OR (user_id IS NULL AND pending_user_id IS NOT NULL));

ALTER TABLE expense_split ADD CONSTRAINT chk_expense_split_user_or_pending 
    CHECK ((user_id IS NOT NULL AND pending_user_id IS NULL) OR (user_id IS NULL AND pending_user_id IS NOT NULL));

CREATE INDEX idx_expense_payments_pending_user ON expense_payments(pending_user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expense_split_pending_user ON expense_split(pending_user_id) WHERE deleted_at IS NULL;

-- +goose Down
ALTER TABLE expense_payments DROP CONSTRAINT IF EXISTS chk_expense_payments_user_or_pending;
ALTER TABLE expense_split DROP CONSTRAINT IF EXISTS chk_expense_split_user_or_pending;
ALTER TABLE expense_payments DROP COLUMN IF EXISTS pending_user_id;
ALTER TABLE expense_split DROP COLUMN IF EXISTS pending_user_id;
```

## Implementation Steps

### 1. Database Queries
**`group_invitations.sql`:**
```sql
-- name: CreateGroupInvitation :one
-- name: GetInvitationByToken :one
-- name: GetInvitationByEmail :one
-- name: ListGroupInvitations :many
-- name: UpdateInvitationStatus :one
-- name: ExpireOldInvitations :exec
```

**`pending_users.sql`:**
```sql
-- name: CreatePendingUser :one
-- name: GetPendingUserByEmail :one
-- name: ConvertPendingUser :exec (link to real user)
```

### 2. Service Layer
**`group_invitation_service.go`:**
- Generate secure tokens
- Send invitation emails
- Handle invitation acceptance
- Create pending user if needed
- Link pending user to real user on signup

**Update `expense_service.go`:**
- Support pending_user_id in payments/splits
- Validate pending users exist
- Auto-convert when pending user signs up

### 3. API Endpoints
```
POST   /groups/{group_id}/invitations        - Send invitation
GET    /groups/{group_id}/invitations        - List invitations
POST   /invitations/{token}/accept           - Accept invitation
POST   /invitations/{token}/decline          - Decline invitation
DELETE /groups/{group_id}/invitations/{id}   - Cancel invitation
```

### 4. Email Integration
- Email template for invitations
- Email service integration (SendGrid, AWS SES, etc.)
- Include group name, inviter name, accept/decline links

### 5. Testing
- Invitation creation and sending
- Token validation
- Acceptance flow
- Pending user creation
- Expense splitting with pending users
- Conversion when pending user signs up

---

# Feature 5: Group Settings & Customization

**Priority:** 🟢 P2  
**Duration:** 2-3 days  
**Dependencies:** None

## Overview
Enhanced group customization with group types, avatars, and default settings.

## Database Changes

### Migration: `enhance_group_settings.sql`
```sql
-- +goose Up
ALTER TABLE groups ADD COLUMN group_type TEXT DEFAULT 'custom' CHECK (group_type IN ('trip', 'household', 'event', 'couple', 'custom'));
ALTER TABLE groups ADD COLUMN avatar_url TEXT;
ALTER TABLE groups ADD COLUMN simplify_debts BOOLEAN DEFAULT true;

-- Update existing settings JSONB to include new fields
-- settings can include: { notifications: true, allow_non_members: false, etc. }

CREATE INDEX idx_groups_type ON groups(group_type) WHERE deleted_at IS NULL;

-- +goose Down
ALTER TABLE groups DROP COLUMN IF EXISTS group_type;
ALTER TABLE groups DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE groups DROP COLUMN IF EXISTS simplify_debts;
```

## Implementation Steps

### 1. Update Group Service
- Add group type selection
- Add avatar upload support (or URL)
- Add simplify_debts preference
- Update settings JSONB handling

### 2. API Endpoints
```
PUT /groups/{id}/settings - Update group settings
PUT /groups/{id}/avatar   - Upload/update avatar
```

### 3. Testing
- Group creation with type
- Settings update
- Avatar upload

---

# Feature 6: Expense Simplification/Optimization

**Priority:** 🟢 P2  
**Duration:** 3-4 days  
**Dependencies:** Balance Calculations (already exists)

## Overview
Optimize debt settlements to minimize number of transactions using graph algorithms.

## Implementation Steps

### 1. Service Layer (`balance_service.go` - enhance)
```go
// Simplified debt structure
type SimplifiedDebt struct {
    From   pgtype.UUID
    To     pgtype.UUID
    Amount string
}

// Optimize debts using greedy algorithm
func (s *balanceService) OptimizeDebts(ctx context.Context, groupID, requesterID pgtype.UUID) ([]SimplifiedDebt, error) {
    // 1. Get all balances
    // 2. Separate creditors (positive) and debtors (negative)
    // 3. Use greedy algorithm to minimize transactions
    // 4. Return optimized payment list
}

// Suggest settlements to clear all debts
func (s *balanceService) SuggestSettlements(ctx context.Context, groupID, requesterID pgtype.UUID) ([]CreateSettlementInput, error)
```

### 2. Algorithm Implementation
```go
// Greedy debt simplification
// Time complexity: O(n log n)
func simplifyDebts(balances []Balance) []SimplifiedDebt {
    creditors := []Balance{} // positive balances
    debtors := []Balance{}   // negative balances
    
    // Sort by absolute amount (descending)
    // Match largest creditor with largest debtor
    // Continue until all balanced
}
```

### 3. API Endpoints
```
GET  /groups/{group_id}/debts/optimized    - Get optimized debt list
POST /groups/{group_id}/debts/settle-all   - Create settlements for all debts
```

### 4. Testing
- Test with various balance scenarios
- Test edge cases (all positive, all negative, balanced)
- Verify transaction count is minimized
- Integration tests

---

# Feature 7: Group Notes (Whiteboard)

**Priority:** 🟢 P2  
**Duration:** 2-3 days  
**Dependencies:** None

## Overview
Shared notes space for groups with edit history tracking.

## Database Changes

### Migration: `create_group_notes.sql`
```sql
-- +goose Up
CREATE TABLE group_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    content TEXT NOT NULL DEFAULT '',
    last_edited_by UUID REFERENCES users(id),
    last_edited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE group_note_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_note_id UUID NOT NULL REFERENCES group_notes(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    edited_by UUID NOT NULL REFERENCES users(id),
    edited_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_group_notes_group_id ON group_notes(group_id);
CREATE INDEX idx_group_note_history_note_id ON group_note_history(group_note_id, edited_at DESC);

-- +goose Down
DROP TABLE IF EXISTS group_note_history;
DROP TABLE IF EXISTS group_notes;
```

## Implementation Steps

### 1. Service Layer (`group_note_service.go`)
- Get/create note for group (one note per group)
- Update note content (save to history)
- Get edit history
- Permission check (group members only)

### 2. API Endpoints
```
GET  /groups/{group_id}/notes          - Get group note
PUT  /groups/{group_id}/notes          - Update note
GET  /groups/{group_id}/notes/history  - Get edit history
```

### 3. Testing
- Note creation and updates
- History tracking
- Permission checks

---

# Feature 8: Group Archiving

**Priority:** 🟢 P2  
**Duration:** 2 days  
**Dependencies:** None

## Overview
Archive groups separately from deletion, preserving data but hiding from active lists.

## Database Changes

### Migration: `add_group_archive.sql`
```sql
-- +goose Up
ALTER TABLE groups ADD COLUMN archived_at TIMESTAMPTZ;
ALTER TABLE groups ADD COLUMN archived_by UUID REFERENCES users(id);

CREATE INDEX idx_groups_archived ON groups(archived_at) WHERE archived_at IS NOT NULL;

-- +goose Down
ALTER TABLE groups DROP COLUMN IF EXISTS archived_at;
ALTER TABLE groups DROP COLUMN IF EXISTS archived_by;
```

## Implementation Steps

### 1. Update Group Service
```go
func (s *groupService) ArchiveGroup(ctx context.Context, groupID, userID pgtype.UUID) error
func (s *groupService) UnarchiveGroup(ctx context.Context, groupID, userID pgtype.UUID) error
func (s *groupService) ListArchivedGroups(ctx context.Context, userID pgtype.UUID) ([]sqlc.Group, error)
```

### 2. Update Queries
- Exclude archived groups from default list
- Add archived groups endpoint

### 3. API Endpoints
```
POST /groups/{id}/archive   - Archive group
POST /groups/{id}/unarchive - Unarchive group
GET  /groups/archived       - List archived groups
```

### 4. Testing
- Archive/unarchive functionality
- Archived groups excluded from lists
- Permission checks (only owner/admin)

---

## Testing Strategy

### Per-Feature Testing
1. **Unit Tests** - Service layer logic
2. **Integration Tests** - Database operations
3. **API Tests** - HTTP endpoints
4. **E2E Tests** - Full user flows

### Test Coverage Goals
- Service layer: >80%
- Repository layer: >70%
- Overall: >75%

---

## Deployment Strategy

### Feature Flags
Use feature flags for gradual rollout:
```go
const (
    FeatureExpenseCategories = "expense_categories"
    FeatureComments         = "comments"
    FeatureActivityFeed     = "activity_feed"
    // ... etc
)
```

### Migration Strategy
1. Run migrations in staging
2. Test with sample data
3. Deploy to production during low-traffic window
4. Monitor for errors
5. Enable feature flags gradually

---

## Success Metrics

### Per Feature
- **Categories**: % of expenses with categories assigned
- **Comments**: Average comments per expense
- **Activity Feed**: Daily active users viewing feed
- **Search**: Search usage rate
- **Invitations**: Invitation acceptance rate
- **Notes**: Groups using notes feature
- **Archive**: Groups archived vs deleted

### Overall
- User engagement increase
- Feature adoption rate
- Bug reports per feature
- Performance impact

---

## Risk Mitigation

### Technical Risks
1. **Performance**: Add indexes, use pagination, monitor query times
2. **Data Migration**: Test migrations thoroughly, have rollback plan
3. **Breaking Changes**: Version API, maintain backward compatibility

### Product Risks
1. **Feature Complexity**: Start with MVP, iterate based on feedback
2. **User Adoption**: Clear documentation, in-app guidance
3. **Scope Creep**: Stick to defined features, defer enhancements to Phase 2

---

## Next Steps

1. ✅ Review and approve this plan
2. [ ] Set up feature flag system
3. [ ] Create feature branches
4. [ ] Start with Feature 1 (Expense Categories)
5. [ ] Weekly progress reviews
6. [ ] Update task.md as features complete

---

## Phase 2 Preview

Features deferred to Phase 2:
- Multi-currency support with exchange rates
- Receipt/image attachments
- Export & reporting
- Spending insights & analytics
- Reminders system
- Payment integrations
- Offline support & sync
- Advanced splitting (itemized bills)

---

## Appendix: Database Schema Summary

### New Tables
1. `expense_categories` - Category definitions
2. `expense_comments` - Comments on expenses
3. `activity_feed` - All group activities
4. `group_invitations` - Email invitations
5. `pending_users` - Non-joined users
6. `group_notes` - Shared group notes
7. `group_note_history` - Note edit history

### Modified Tables
1. `expenses` - Add category_id, tags
2. `groups` - Add group_type, avatar_url, simplify_debts, archived_at
3. `expense_payments` - Add pending_user_id
4. `expense_split` - Add pending_user_id

### Total New Columns: ~15
### Total New Tables: 7
### Total New Indexes: ~25
