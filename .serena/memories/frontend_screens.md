# Split+ Frontend Screen Inventory

Based on the backend API structure, here is a comprehensive list of all screens needed for a functional Split+ expense splitting app.

## Core Navigation Structure

```
Layout (Authenticated)
├── Sidebar/Navigation
│   ├── Dashboard (Home)
│   ├── Groups
│   ├── Friends
│   ├── Activity
│   └── Settings
└── Main Content Area

Public Routes (Unauthenticated)
├── Login
├── Register
└── Accept Invitation
```

---

## 1. Authentication Screens

### 1.1 Login Screen
- **Route**: `/login`
- **Backend API**: `POST /auth/login`, `POST /auth/refresh`
- **Features**:
  - Email and password input
  - "Remember me" option
  - Forgot password link
  - Register link
  - OAuth buttons (if supported)
- **Data Needed**: LoginRequest, LoginResponse

### 1.2 Register/Create Account Screen
- **Route**: `/register`
- **Backend API**: `POST /users`
- **Features**:
  - Name, email, password fields
  - Password strength indicator
  - Terms of service checkbox
  - Login link for existing users
- **Data Needed**: CreateUserRequest, UserResponse

---

## 2. Dashboard/Home Screen

### 2.1 Dashboard
- **Route**: `/` or `/dashboard`
- **Backend API**: `GET /users/me/balances`
- **Features**:
  - Overall balance summary (what you owe vs. owed to you)
  - Quick stats: total groups, pending invitations, recent activity
  - Recent expenses across all groups
  - Quick actions: Add expense, Create group, Invite friend
- **Components**:
  - Balance summary cards
  - Recent activity feed
  - Quick action buttons
- **Data Needed**: OverallUserBalanceResponse, recent activities

---

## 3. Group Management Screens

### 3.1 Groups List Screen
- **Route**: `/groups`
- **Backend API**: `GET /groups`
- **Features**:
  - List of all user's groups
  - Group cards with name, member count, balance preview
  - Create new group button
  - Search/filter groups
- **Data Needed**: UserGroupResponse[]

### 3.2 Group Detail Screen
- **Route**: `/groups/:groupId`
- **Backend API**: 
  - `GET /groups/:groupId/members`
  - `GET /groups/:groupId/expenses`
  - `GET /groups/:groupId/balances`
  - `GET /groups/:groupId/debts`
- **Features**:
  - Group header (name, description, currency)
  - Tab navigation:
    - **Expenses**: List of all expenses with search
    - **Balances**: Who owes whom in the group
    - **Members**: Group members list
    - **Activity**: Group activity feed
    - **Settings**: Group settings (for admins)
  - FAB (Floating Action Button) for adding expenses
- **Components**:
  - Expense list with filters
  - Balance visualization
  - Member avatars
  - Activity timeline

### 3.3 Create Group Screen
- **Route**: `/groups/new`
- **Backend API**: `POST /groups`
- **Features**:
  - Group name input
  - Description (optional)
  - Default currency selection
  - Default split method selection
  - Create button
- **Data Needed**: CreateGroupRequest, CreateGroupResponse

### 3.4 Invite Member Screen
- **Route**: `/groups/:groupId/invite`
- **Backend API**: `POST /groups/:groupId/invitations`
- **Features**:
  - Email input for invitation
  - Role selection (member/admin)
  - Generate shareable link option
  - List of pending invitations
- **Data Needed**: CreateInvitationRequest

### 3.5 Group Settings Screen
- **Route**: `/groups/:groupId/settings`
- **Backend API**: Various group update endpoints
- **Features**:
  - Edit group name/description
  - Change default currency
  - Change default split method
  - Leave group
  - Delete group (owner only)
  - Manage members (remove/promote)

---

## 4. Expense Management Screens

### 4.1 Expense List (Within Group)
- **Route**: `/groups/:groupId/expenses`
- **Backend API**: 
  - `GET /groups/:groupId/expenses`
  - `GET /groups/:groupId/expenses/search`
- **Features**:
  - Search bar for expenses
  - Filter by date range, category, member
  - Sort by date, amount
  - Expense cards with:
    - Title, amount, date
    - Paid by info
    - Split preview
    - Category icon
- **Data Needed**: GroupExpenseResponse[]

### 4.2 Add Expense Screen
- **Route**: `/groups/:groupId/expenses/new`
- **Backend API**: `POST /groups/:groupId/expenses`
- **Features**:
  - Description/title input
  - Amount input with currency
  - Date picker
  - Paid by selection (who paid)
  - Split method selection:
    - Equal split
    - Unequal (by amount)
    - Percentage
    - Shares
  - Member selection for split
  - Category selection
  - Tags input
  - Notes/attachments
  - Receipt photo upload
- **Data Needed**: CreateExpenseRequest (PaymentRequest[], SplitRequest[])

### 4.3 Expense Detail Screen
- **Route**: `/groups/:groupId/expenses/:expenseId`
- **Backend API**: 
  - `GET /groups/:groupId/expenses/:expenseId`
  - `GET /groups/:groupId/expenses/:expenseId/comments`
  - `GET /groups/:groupId/expenses/:expenseId/history`
- **Features**:
  - Expense details card
  - Payment breakdown (who paid how much)
  - Split breakdown (who owes what)
  - Comments section (add/view/delete)
  - Activity/history log
  - Edit and delete buttons (for creator)
  - Settle button
- **Components**:
  - Expense card
  - Payment list
  - Split visualization
  - Comment thread
  - History timeline
- **Data Needed**: ExpenseResponse, ExpenseComment[], GroupActivity[]

### 4.4 Edit Expense Screen
- **Route**: `/groups/:groupId/expenses/:expenseId/edit`
- **Backend API**: `PUT /groups/:groupId/expenses/:expenseId`
- **Features**: Same as Add Expense but pre-populated
- **Data Needed**: UpdateExpenseRequest

---

## 5. Settlement Screens

### 5.1 Settlements List (Group)
- **Route**: `/groups/:groupId/settlements`
- **Backend API**: `GET /groups/:groupId/settlements`
- **Features**:
  - List of all settlements in the group
  - Filter by status (pending, completed)
  - Settlement cards with payer, payee, amount, status
- **Data Needed**: SettlementResponse[]

### 5.2 Record Settlement Screen
- **Route**: `/groups/:groupId/settlements/new`
- **Backend API**: `POST /groups/:groupId/settlements`
- **Features**:
  - Select payer (who paid)
  - Select payee (who received)
  - Amount input
  - Payment method selection
  - Transaction reference (optional)
  - Notes
  - Date picker
- **Data Needed**: CreateSettlementRequest

### 5.3 Settlement Detail Screen
- **Route**: `/groups/:groupId/settlements/:settlementId`
- **Backend API**: 
  - `GET /groups/:groupId/settlements/:settlementId`
  - `PATCH /groups/:groupId/settlements/:settlementId/status`
- **Features**:
  - Settlement details
  - Status badge
  - Update status (confirm receipt)
  - Edit/Delete options
- **Data Needed**: SettlementResponse, UpdateSettlementStatusRequest

---

## 6. Balance & Debt Screens

### 6.1 Group Balances Screen
- **Route**: `/groups/:groupId/balances`
- **Backend API**: 
  - `GET /groups/:groupId/balances`
  - `GET /groups/:groupId/debts`
- **Features**:
  - Simplified debt view (who owes whom)
  - Individual balance cards
  - Visual debt graph/network
  - "Settle Up" quick actions
- **Components**:
  - Balance cards
  - Debt visualization (graph/chart)
  - Settlement suggestions
- **Data Needed**: BalanceResponse[], SimplifiedDebtResponse

---

## 7. Recurring Expense Screens

### 7.1 Recurring Expenses List
- **Route**: `/groups/:groupId/recurring`
- **Backend API**: `GET /groups/:groupId/recurring-expenses`
- **Features**:
  - List of recurring expenses
  - Status indicators (active/paused)
  - Next occurrence date
  - Quick actions: Edit, Delete, Pause, Generate now
- **Data Needed**: RecurringExpenseResponse[]

### 7.2 Add/Edit Recurring Expense Screen
- **Route**: `/groups/:groupId/recurring/new` or `/groups/:groupId/recurring/:id/edit`
- **Backend API**: 
  - `POST /groups/:groupId/recurring-expenses`
  - `PUT /groups/:groupId/recurring-expenses/:id`
- **Features**:
  - Same as regular expense +:
  - Repeat interval (weekly, monthly, yearly)
  - Day of month/week selection
  - Start date
  - End date (optional)
  - Is active toggle
- **Data Needed**: CreateRecurringExpenseRequest, UpdateRecurringExpenseRequest

### 7.3 Recurring Expense Detail Screen
- **Route**: `/groups/:groupId/recurring/:id`
- **Backend API**: 
  - `GET /groups/:groupId/recurring-expenses/:id`
  - `POST /groups/:groupId/recurring-expenses/:id/generate`
- **Features**:
  - Recurring expense details
  - Schedule visualization
  - Generated expenses history
  - Manual generate button
  - Edit/Delete options
- **Data Needed**: RecurringExpenseResponse

---

## 8. Category Management Screens

### 8.1 Categories List
- **Route**: `/groups/:groupId/categories`
- **Backend API**: `GET /groups/:groupId/categories`
- **Features**:
  - List of group categories
  - Preset categories to add
  - Custom category creation
  - Edit/Delete custom categories
- **Data Needed**: ExpenseCategoryResponse[]

### 8.2 Add Category Screen
- **Route**: `/groups/:groupId/categories/new`
- **Backend API**: `POST /groups/:groupId/categories`
- **Features**:
  - Name input
  - Icon selection
  - Color picker
  - Preset categories gallery
- **Data Needed**: CreateCategoryRequest

---

## 9. Friends Management Screens

### 9.1 Friends List Screen
- **Route**: `/friends`
- **Backend API**: `GET /friends`
- **Features**:
  - List of friends
  - Balance with each friend
  - Quick actions: Add expense, Settle up, Remove friend
  - Search friends
- **Data Needed**: FriendResponse[]

### 9.2 Friend Requests Screen
- **Route**: `/friends/requests`
- **Backend API**: 
  - `GET /friends/requests/incoming`
  - `GET /friends/requests/outgoing`
- **Features**:
  - Incoming requests with accept/decline
  - Outgoing requests with cancel option
- **Data Needed**: FriendRequestResponse[]

### 9.3 Add Friend Screen
- **Route**: `/friends/add`
- **Backend API**: `POST /friends/requests`
- **Features**:
  - Search by email or name
  - Send friend request
  - Share invite link
- **Data Needed**: SendFriendRequestRequest

### 9.4 Friend Detail Screen
- **Route**: `/friends/:friendId`
- **Backend API**: 
  - `GET /friends/:friendId/expenses`
  - `GET /friends/:friendId/settlements`
- **Features**:
  - Friend profile/info
  - Expense history with friend
  - Settlement history
  - Balance with friend
  - Add expense button
  - Settle up button
  - Remove friend option
- **Components**:
  - Expense list (friend-specific)
  - Settlement list
  - Balance summary

---

## 10. Activity & History Screens

### 10.1 Group Activity Screen
- **Route**: `/groups/:groupId/activity`
- **Backend API**: `GET /groups/:groupId/activity`
- **Features**:
  - Timeline of all group activities
  - Filter by activity type (expense, settlement, member changes)
  - Infinite scroll
- **Data Needed**: GroupActivityResponse[]

### 10.2 Global Activity Screen
- **Route**: `/activity`
- **Backend API**: Aggregated from multiple sources
- **Features**:
  - All activities across groups and friends
  - Filter by type
  - Group by date
- **Data Needed**: ActivityResponse[]

---

## 11. Invitation Screens

### 11.1 Accept Invitation Screen
- **Route**: `/invitations/:token`
- **Backend API**: 
  - `GET /invitations/:token`
  - `POST /invitations/:token/accept` (authenticated)
  - `POST /invitations/:token/join` (smart join)
- **Features**:
  - Show invitation details (group name, inviter)
  - If logged in: Join button
  - If not logged in: Register/Login and join flow
- **Data Needed**: InvitationResponse

---

## 12. User Settings Screens

### 12.1 Profile Settings
- **Route**: `/settings/profile`
- **Backend API**: User update endpoints
- **Features**:
  - Edit name, email
  - Change password
  - Upload avatar
  - Language/Timezone settings
- **Data Needed**: UserResponse

### 12.2 Notification Settings
- **Route**: `/settings/notifications`
- **Backend API**: Settings update
- **Features**:
  - Email notification toggles
  - Push notification settings
  - Activity digest preferences

### 12.3 Preferences
- **Route**: `/settings/preferences`
- **Features**:
  - Default currency
  - Default split method
  - Date format
  - Number format

---

## 13. Modal Screens / Overlays

### 13.1 Confirmation Modals
- Delete confirmation
- Leave group confirmation
- Remove friend confirmation

### 13.2 Quick Actions
- Quick add expense modal
- Quick settle up modal
- Share group link modal

### 13.3 Selectors
- Date range picker
- Member multi-select
- Category selector
- Currency selector
- Split calculator

---

## Screen Summary Count

| Category | Screen Count |
|----------|--------------|
| Authentication | 2 |
| Dashboard | 1 |
| Groups | 5 |
| Expenses | 4 |
| Settlements | 3 |
| Balances | 1 |
| Recurring Expenses | 3 |
| Categories | 2 |
| Friends | 4 |
| Activity | 2 |
| Invitations | 1 |
| Settings | 3 |
| **Total** | **31 screens** |

---

## Navigation Hierarchy

```
App
├── PublicLayout
│   ├── /login
│   ├── /register
│   └── /invitations/:token
│
└── AuthenticatedLayout (with Sidebar)
    ├── /dashboard
    ├── /groups
    │   ├── /groups/new
    │   └── /groups/:groupId
    │       ├── /groups/:groupId/expenses
    │       │   ├── /groups/:groupId/expenses/new
    │       │   └── /groups/:groupId/expenses/:expenseId
    │       │       └── /groups/:groupId/expenses/:expenseId/edit
    │       ├── /groups/:groupId/balances
    │       ├── /groups/:groupId/settlements
    │       │   └── /groups/:groupId/settlements/new
    │       ├── /groups/:groupId/members
    │       │   └── /groups/:groupId/invite
    │       ├── /groups/:groupId/activity
    │       ├── /groups/:groupId/recurring
    │       │   ├── /groups/:groupId/recurring/new
    │       │   └── /groups/:groupId/recurring/:id
    │       ├── /groups/:groupId/categories
    │       │   └── /groups/:groupId/categories/new
    │       └── /groups/:groupId/settings
    ├── /friends
    │   ├── /friends/add
    │   ├── /friends/requests
    │   └── /friends/:friendId
    ├── /activity
    └── /settings
        ├── /settings/profile
        ├── /settings/notifications
        └── /settings/preferences
```

---

## Key UI Components Needed

Based on these screens, the following reusable components should be built:

1. **Layout Components**
   - AuthenticatedLayout (with sidebar)
   - PublicLayout

2. **Navigation Components**
   - Sidebar
   - TabNavigation
   - Breadcrumbs

3. **Data Display Components**
   - GroupCard
   - ExpenseCard
   - BalanceCard
   - ActivityItem
   - MemberAvatar
   - SettlementCard

4. **Form Components**
   - ExpenseForm
   - SplitCalculator
   - CategorySelector
   - MemberMultiSelect
   - CurrencyInput

5. **Modal Components**
   - ConfirmationModal
   - InviteModal
   - QuickAddModal

6. **Chart Components**
   - BalanceChart
   - DebtNetworkGraph
   - ExpenseBreakdown

This inventory provides a complete roadmap for building the Split+ frontend to support all backend functionality.
